package wallet

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
    "github.com/Luckyisrael/solana-wallet-api/internal/api/dto"
    "github.com/Luckyisrael/solana-wallet-api/internal/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCreateWallet_Custodial(t *testing.T) {
	os.Setenv("MASTER_KEY", "4v9mK9jZp8xLmN7qRt5sWx2cYb3aUe1fGh7jKi8nPo2=")
	defer os.Unsetenv("MASTER_KEY")

	// Use in-memory 
	// mock repo
	mockRepo := &repo.WalletRepo{} // stub

	svc := NewService(mockRepo)
	req := &dto.CreateWalletRequest{ReturnPrivateKey: false}

	resp, err := svc.CreateWallet(context.Background(), req)
	assert.NoError(t, err)
	custodial := resp.(*dto.WalletAddressOnly)
	assert.NotEmpty(t, custodial.Address)
}

func TestCreateWallet_NonCustodial(t *testing.T) {
	os.Setenv("MASTER_KEY", "4v9mK9jZp8xLmN7qRt5sWx2cYb3aUe1fGh7jKi8nPo2=")
	defer os.Unsetenv("MASTER_KEY")

	mockRepo := &repo.WalletRepo{}
	svc := NewService(mockRepo)
	req := &dto.CreateWalletRequest{ReturnPrivateKey: true}

	resp, err := svc.CreateWallet(context.Background(), req)
	assert.NoError(t, err)
	full := resp.(*dto.WalletWithKey)
	assert.NotEmpty(t, full.Address)
	assert.NotEmpty(t, full.PrivateKey)
	assert.Equal(t, 88, len(full.PrivateKey)) // base58 length
}