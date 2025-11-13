package repo

import "time"

type Wallet struct {
	ID           int       `db:"id"`
	Address      string    `db:"address"`
	EncryptedKey []byte    `db:"encrypted_key"` // AES-256-GCM
	IV           []byte    `db:"iv"`            // 12 bytes
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}