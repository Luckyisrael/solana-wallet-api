package dto

type TransferRequest struct {
	FromAddress string `json:"fromAddress" binding:"required" example:"7x9kLmN2pQ...abc123"`
	ToAddress   string `json:"toAddress" binding:"required" example:"9xY8kLmN2pQ...xyz789"`
	Amount      string `json:"amount" binding:"required" example:"0.1"` // formated amount
	Mint        string `json:"mint,omitempty" example:"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyK9u7u"` // SPL Tokens
}

type TransferResponse struct {
	Signature     string `json:"signature" example:"5e9x...abc123"`
	SignedTxBase64 string `json:"signedTxBase64" example:"AQAA...=="`
	FeePayer      string `json:"feePayer" example:"7x9kLmN2pQ...abc123"`
}