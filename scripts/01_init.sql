CREATE TABLE IF NOT EXISTS wallets (
    id SERIAL PRIMARY KEY,
    address TEXT UNIQUE NOT NULL,
    encrypted_key BYTEA NOT NULL,
    iv BYTEA NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallets_address ON wallets(address);
