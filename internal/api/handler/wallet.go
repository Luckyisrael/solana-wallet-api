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
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
            return
        }

        resp, err := walletService.CreateWallet(c.Request.Context(), &req)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        if req.ReturnPrivateKey {
            c.JSON(http.StatusCreated, resp.(*dto.WalletWithKey))
        } else {
            c.JSON(http.StatusCreated, resp.(*dto.WalletAddressOnly))
        }
    }
}
