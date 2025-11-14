package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/Luckyisrael/solana-wallet-api/internal/api/dto"
    "github.com/Luckyisrael/solana-wallet-api/internal/service/wallet"
)

// CreateWallet godoc
// @Summary      Create a new Solana wallet
// @Description  Generates a new keypair. If returnPrivateKey=true, includes private key (one-time only).
// @Tags         wallets
// @Accept       json
// @Produce      json
// @Param        body  body  dto.CreateWalletRequest  true  "Request body"
// @Success      201   {object}  dto.WalletAddressOnly
// @Success      201   {object}  dto.WalletWithKey
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /wallets [post]
func CreateWallet(walletService *wallet.Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req dto.CreateWalletRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            code := http.StatusBadRequest
            c.JSON(code, gin.H{"success": false, "data": nil, "message": "invalid request", "responseCode": code})
            return
        }

        resp, err := walletService.CreateWallet(c.Request.Context(), &req)
        if err != nil {
            code := http.StatusInternalServerError
            c.JSON(code, gin.H{"success": false, "data": nil, "message": err.Error(), "responseCode": code})
            return
        }
        code := http.StatusCreated
        c.JSON(code, gin.H{"success": true, "data": resp, "message": "", "responseCode": code})
    }
}
