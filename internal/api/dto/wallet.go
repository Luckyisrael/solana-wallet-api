package dto

// CreateWalletRequest - input
type CreateWalletRequest struct {
	ReturnPrivateKey bool `json:"returnPrivateKey" example:"true" default:"false"`
	ExportSeedPhrase bool `json:"exportSeedPhrase" example:"false" default:"false"` // for future implementation if i have time to update thihs
}

// WalletAddressOnly - custodial response
type WalletAddressOnly struct {
	Address string `json:"address" example:"4gqrwNqUmc4afEb1iFEX8VjEjWJNfnmdMbkcdsUr9eRY"`
}

// WalletWithKey - non-custodial response
type WalletWithKey struct {
	Address    string `json:"address" example:"9xY8kLmN2pQ...abc123"`
	SeedPhrase string `json:"seedPhrase" example:"abandon ... zoo"` // 12 words
}