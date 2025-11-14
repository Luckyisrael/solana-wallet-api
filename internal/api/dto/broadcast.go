package dto

type BroadcastRequest struct {
    SignedTxBase64   string `json:"signedTxBase64" binding:"required" example:"AQAA...=="`
    IdempotencyKey   string `json:"idempotencyKey" example:"tx-12345"`
}

type BroadcastResponse struct {
	Signature string `json:"signature" example:"5e9x...abc123"`
	Status    string `json:"status" example:"confirmed"` // pending, confirmed, failed
	Slot      uint64 `json:"slot,omitempty"`
	Error     string `json:"error,omitempty"`
}
