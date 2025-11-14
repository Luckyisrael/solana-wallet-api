CREATE TABLE IF NOT EXISTS api_keys (
    id SERIAL PRIMARY KEY,
    key_hash TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    rate_limit_per_min INT DEFAULT 100
);

-- Insert first key (run once)
INSERT INTO api_keys (key_hash, name) VALUES
('$2a$10$4x9kLmN2pQ7xY8kLmN2pQ7xY8kLmN2pQ7xY8kLmN2pQ7xY8kLmN2pQ', 'dev-key-1')
ON CONFLICT DO NOTHING;