package transfer

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/Luckyisrael/solana-wallet-api/internal/api/dto"
	"github.com/Luckyisrael/solana-wallet-api/internal/crypto"
	"github.com/Luckyisrael/solana-wallet-api/internal/redis"
	"github.com/Luckyisrael/solana-wallet-api/internal/repo"
	"go.uber.org/zap"
)

const (
	rateLimitKey = "rate:transfer:%s"
	rateLimitMax = 1
	rateLimitTTL = time.Second
)

type Service struct {
	walletRepo   *repo.WalletRepo
	solanaClient *solana.Client
	redisClient  *redis.Client
	masterKey    string
	logger       *zap.Logger
}

func NewService(walletRepo *repo.WalletRepo, solanaClient *solana.Client, redisClient *redis.Client, masterKey string) *Service {
	logger, _ := zap.NewProduction()
	return &Service{
		walletRepo:   walletRepo,
		solanaClient: solanaClient,
		redisClient:  redisClient,
		masterKey:    masterKey,
		logger:       logger,
	}
}

func (s *Service) Transfer(ctx context.Context, req *dto.TransferRequest) (*dto.TransferResponse, error) {
	fromAddr, err := solana.PublicKeyFromBase58(req.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid from address")
	}

	toAddr, err := solana.PublicKeyFromBase58(req.ToAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid to address")
	}

	// Rate limiting
	if err := s.rateLimit(ctx, req.FromAddress); err != nil {
		return nil, err
	}

	// Get encrypted key
	wallet, err := s.walletRepo.GetByAddress(ctx, req.FromAddress)
	if err != nil || wallet == nil {
		return nil, fmt.Errorf("wallet not found")
	}

	// Decrypt private key
	privateKey, err := crypto.Decrypt(wallet.EncryptedKey, s.masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key")
	}
	signer := solana.PrivateKey(privateKey)

	// Get latest blockhash
	blockhash, err := s.solanaClient.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, err
	}

	var signedTx *solana.Transaction
	if req.Mint == "" {
		// SOL transfer
		lamports, _ := parseSOLAmount(req.Amount)
		signedTx, err = s.buildSOLTransfer(ctx, fromAddr, toAddr, lamports, blockhash, signer)
	} else {
		// SPL token transfer
		mint, _ := solana.PublicKeyFromBase58(req.Mint)
		amount, _ := parseTokenAmount(req.Amount, 6) // default 6 decimals
		signedTx, err = s.buildTokenTransfer(ctx, fromAddr, toAddr, mint, amount, blockhash, signer)
	}
	if err != nil {
		return nil, err
	}

	// Serialize
	txBytes, err := signedTx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	sig := signedTx.Signatures[0].String()
	base64Tx := base64.StdEncoding.EncodeToString(txBytes)

	s.logger.Info("Transfer built",
		zap.String("from", req.FromAddress),
		zap.String("to", req.ToAddress),
		zap.String("signature", sig),
	)

	return &dto.TransferResponse{
		Signature:     sig,
		SignedTxBase64: base64Tx,
		FeePayer:      fromAddr.String(),
	}, nil
}

func (s *Service) rateLimit(ctx context.Context, address string) error {
	key := fmt.Sprintf(rateLimitKey, address)
	count, err := s.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		s.redisClient.Expire(ctx, key, rateLimitTTL)
	}
	if count > rateLimitMax {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

func (s *Service) buildSOLTransfer(ctx context.Context, from, to solana.PublicKey, lamports uint64, blockhash solana.Hash, signer solana.PrivateKey) (*solana.Transaction, error) {
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				lamports,
				from,
				to,
			).Build(),
		},
		blockhash,
		solana.TransactionPayer(from),
	)
	if err != nil {
		return nil, err
	}
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(signer.PublicKey()) {
			return &signer
		}
		return nil
	})
	return tx, err
}

func (s *Service) buildTokenTransfer(ctx context.Context, from, to, mint solana.PublicKey, amount uint64, blockhash solana.Hash, signer solana.PrivateKey) (*solana.Transaction, error) {
	// Get source token account
	srcAcc, err := s.getOrCreateTokenAccount(ctx, from, mint)
	if err != nil {
		return nil, err
	}

	// Get or create destination token account
	dstAcc, err := s.getOrCreateTokenAccount(ctx, to, mint)
	if err != nil {
		return nil, err
	}

	instructions := []solana.Instruction{
		token.NewTransferInstruction(
			amount,
			srcAcc,
			dstAcc,
			from,
		).Build(),
	}

	tx, err := solana.NewTransaction(instructions, blockhash, solana.TransactionPayer(from))
	if err != nil {
		return nil, err
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(signer.PublicKey()) {
			return &signer
		}
		return nil
	})
	return tx, err
}

func (s *Service) getOrCreateTokenAccount(ctx context.Context, owner, mint solana.PublicKey) (solana.PublicKey, error) {
	accounts, err := s.solanaClient.GetTokenAccountsByOwner(ctx, owner, &rpc.GetTokenAccountsConfig{Mint: &mint}, nil)
	if err != nil {
		return solana.PublicKey{}, err
	}
	if len(accounts.Value) > 0 {
		return accounts.Value[0].PublicKey, nil
	}

	// Create associated token account
	ata, _, err := solana.FindAssociatedTokenAddress(owner, mint)
	if err != nil {
		return solana.PublicKey{}, err
	}
	return ata, nil
}

func parseSOLAmount(amountStr string) (uint64, error) {
	f, _ := new(big.Float).SetString(amountStr)
	f.Mul(f, big.NewFloat(1e9))
	i, _ := f.Uint64()
	return i, nil
}

func parseTokenAmount(amountStr string, decimals uint8) (uint64, error) {
	f, _ := new(big.Float).SetString(amountStr)
	scale := new(big.Float).SetFloat64(float64(1))
	for i := uint8(0); i < decimals; i++ {
		scale.Mul(scale, big.NewFloat(10))
	}
	f.Mul(f, scale)
	i, _ := f.Uint64()
	return i, nil
}