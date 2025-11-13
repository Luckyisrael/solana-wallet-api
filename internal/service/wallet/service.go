package wallet

import (
    "context"
    "os"

    "github.com/gagliardetto/solana-go"
    "github.com/Luckyisrael/solana-wallet-api/internal/api/dto"
    "github.com/Luckyisrael/solana-wallet-api/internal/crypto"
    "github.com/Luckyisrael/solana-wallet-api/internal/repo"

    "github.com/tyler-smith/go-bip39"
    "golang.org/x/crypto/ed25519"
)

type Service struct {
    repo      *repo.WalletRepo
    masterKey string
}

func NewService(repo *repo.WalletRepo) *Service {
	masterKey := os.Getenv("MASTER_KEY")
	if masterKey == "" {
		panic("MASTER_KEY is required")
	}

    return &Service{
        repo:      repo,
        masterKey: masterKey,
    }
}

func (s *Service) CreateWallet(ctx context.Context, req *dto.CreateWalletRequest) (any, error) {
    var seedPhrase string
    var priv ed25519.PrivateKey

    if req.ReturnPrivateKey {
        entropy, err := bip39.NewEntropy(128)
        if err != nil {
            return nil, err
        }
        seedPhrase, err = bip39.NewMnemonic(entropy)
        if err != nil {
            return nil, err
        }
        seed := bip39.NewSeed(seedPhrase, "")
        priv = ed25519.NewKeyFromSeed(seed[:32])
    } else {
        w := solana.NewWallet()
        priv = ed25519.PrivateKey(w.PrivateKey)
    }

    pub := priv.Public().(ed25519.PublicKey)
    address := solana.PublicKeyFromBytes(pub).String()

    encryptedKey, iv, err := crypto.Encrypt(priv, s.masterKey)
    if err != nil {
        return nil, err
    }
    dbWallet := &repo.Wallet{
        Address:      address,
        EncryptedKey: encryptedKey,
        IV:           iv,
    }
    if err := s.repo.Create(ctx, dbWallet); err != nil {
        return nil, err
    }

    if req.ReturnPrivateKey {
        return &dto.WalletWithKey{Address: address, SeedPhrase: seedPhrase}, nil
    }
    return &dto.WalletAddressOnly{Address: address}, nil
}
