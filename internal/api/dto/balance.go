package dto

type TokenBalance struct {
	Mint     string  `json:"mint" example:"So11111111111111111111111111111111111111112"`
	Symbol   string  `json:"symbol" example:"SOL"`
	Name     string  `json:"name" example:"Wrapped SOL"`
	Amount   string  `json:"amount" example:"1.500000"` 
	Decimals uint8   `json:"decimals" example:"9"`
	UIAmount float64 `json:"uiAmount" example:"1.5"` // float for UI
}

type BalanceResponse struct {
	Address string `json:"address" example:"7x9kLmN2pQ...abc123"`
	SOL     string `json:"sol" example:"1.500000"` // lamports → SOL
	Tokens  []TokenBalance `json:"tokens"`
}