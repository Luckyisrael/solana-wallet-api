package repo

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type APIKey struct {
	ID             int
	KeyHash        string
	Name           string
	IsActive       bool
	RateLimitPerMin int
}

type APIKeyRepo struct {
	db *pgxpool.Pool
}

func NewAPIKeyRepo(db *pgxpool.Pool) *APIKeyRepo {
	return &APIKeyRepo{db: db}
}

func (r *APIKeyRepo) ValidateKey(ctx context.Context, apiKey string) (*APIKey, error) {
	var key APIKey
	query := `SELECT id, key_hash, name, is_active, rate_limit_per_min FROM api_keys WHERE key_hash = $1`
	err := r.db.QueryRow(ctx, query, hashAPIKey(apiKey)).Scan(
		&key.ID, &key.KeyHash, &key.Name, &key.IsActive, &key.RateLimitPerMin,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !key.IsActive {
		return nil, nil
	}
	return &key, nil
}

func hashAPIKey(key string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(key), 10)
	return string(h)
}