package transfer

import (
    "context"
    "encoding/base64"
    "fmt"
    "math/big"
    "time"

    solanago "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/programs/system"

    "github.com/Luckyisrael/solana-wallet-api/internal/api/dto"
    "github.com/Luckyisrael/solana-wallet-api/internal/crypto"
    "github.com/Luckyisrael/solana-wallet-api/internal/redis"
    "github.com/Luckyisrael/solana-wallet-api/internal/repo"
    "github.com/Luckyisrael/solana-wallet-api/internal/solana"
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
    fromAddr, err := solanago.PublicKeyFromBase58(req.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid from address")
	}

    toAddr, err := solanago.PublicKeyFromBase58(req.ToAddress)
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
    signer := solanago.PrivateKey(privateKey)

    // Get latest blockhash
    blockhash, err := s.solanaClient.GetLatestBlockhash(ctx)
    if err != nil {
        return nil, err
    }

    lamports, err := parseSOLAmount(req.Amount)
    if err != nil {
        return nil, err
    }
    signedTx, err := s.buildSOLTransfer(ctx, fromAddr, toAddr, lamports, blockhash, signer)
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

func (s *Service) buildSOLTransfer(ctx context.Context, from, to solanago.PublicKey, lamports uint64, blockhash solanago.Hash, signer solanago.PrivateKey) (*solanago.Transaction, error) {
    tx, err := solanago.NewTransaction(
        []solanago.Instruction{
            system.NewTransferInstruction(
                lamports,
                from,
                to,
            ).Build(),
        },
        blockhash,
        solanago.TransactionPayer(from),
    )
    if err != nil {
        return nil, err
    }
    _, err = tx.Sign(func(key solanago.PublicKey) *solanago.PrivateKey {
        if key.Equals(signer.PublicKey()) {
            return &signer
        }
        return nil
    })
    return tx, err
}


func parseSOLAmount(amountStr string) (uint64, error) {
    f, ok := new(big.Float).SetString(amountStr)
    if !ok {
        return 0, fmt.Errorf("invalid amount")
    }
    f.Mul(f, big.NewFloat(1e9))
    i, _ := f.Uint64()
    if i == 0 {
        return 0, fmt.Errorf("amount must be greater than 0")
    }
    return i, nil
}

// SPL token transfer removed
