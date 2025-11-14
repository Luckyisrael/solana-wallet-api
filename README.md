# Solana Wallet API

Production-ready Solana wallet backend that supports wallet creation, SOL balance retrieval, SOL transfers, and transaction broadcast.

## Prerequisites
- Go 1.21+
- PostgreSQL (default: `127.0.0.1:5433`, database `wallet`)
- Redis (default: `localhost:6379`)

## Configuration
- Environment:
  - `MASTER_KEY` (required): passphrase used to encrypt private keys (AES-256-GCM)
- Config file: `internal/config/config.yaml` (overrides via env supported)

## Run
1. Start Postgres and Redis (Docker Compose available in `docker/docker-compose.yml`).
2. Set `MASTER_KEY` in your environment.
3. Start the API:
   - `go run cmd/api/main.go`

Server listens on `:8080` by default.

## Endpoints
- `POST /v1/wallets` — Create a new wallet
  - Body: `{ "returnPrivateKey": true|false }`
  - Response: address; when `returnPrivateKey=true`, includes seed phrase

- `GET /v1/wallets/:address/balance` — Get SOL balance

- `POST /v1/transactions/transfer` — Transfer SOL
  - Body: `{ "fromAddress": string, "toAddress": string, "amount": "0.1" }`
  - Returns signed transaction base64 and signature

- `POST /v1/transactions/broadcast` — Broadcast a signed transaction
  - Body: `{ "signedTxBase64": string, "idempotencyKey"?: string }`
  - `idempotencyKey` is optional; if omitted, a SHA-256 of the base64 payload is used
  - Returns signature, status, and slot

- `GET /v1/transactions/:address/history` — Transaction history
  - Query: `limit` (optional, default 20), `before` (optional signature cursor)

## Notes
- No API key authentication is required.
- Private keys are encrypted using AES-256-GCM with scrypt-derived keys.

## Response Format
All endpoints return: `{ "success": boolean, "data": any, "message": string, "responseCode": number }`
