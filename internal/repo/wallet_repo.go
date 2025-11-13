package repo

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepo struct {
	db *pgxpool.Pool
}

func NewWalletRepo(db *pgxpool.Pool) *WalletRepo {
	return &WalletRepo{db: db}
}

func (r *WalletRepo) Create(ctx context.Context, w *Wallet) error {
	query := `
		INSERT INTO wallets (address, encrypted_key, iv, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, w.Address, w.EncryptedKey, w.IV)
	return err
}

func (r *WalletRepo) GetByAddress(ctx context.Context, address string) (*Wallet, error) {
	w := &Wallet{}
	query := `SELECT id, address, encrypted_key, iv, created_at, updated_at FROM wallets WHERE address = $1`
	err := r.db.QueryRow(ctx, query, address).Scan(
		&w.ID, &w.Address, &w.EncryptedKey, &w.IV, &w.CreatedAt, &w.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return w, nil
}