package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Luckyisrael/solana-wallet-api/internal/service/balance"
)

// GetBalance godoc
// @Summary      Get wallet balance
// @Description  Returns SOL and SPL token balances with. Cached for 5 minutes.
// @Tags         wallets
// @Produce      json
// @Param        address  path  string  true  "Wallet address"
// @Success      200  {object}  dto.BalanceResponse
// @Failure      400  {object}  map[string]string
// @Failure      429  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /wallets/{address}/balance [get]
func GetBalance(balanceService *balance.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
        if address == "" {
            code := http.StatusBadRequest
            c.JSON(code, gin.H{"success": false, "data": nil, "message": "address is required", "responseCode": code})
            return
        }

        resp, err := balanceService.GetBalance(c.Request.Context(), address)
        if err != nil {
            if err.Error() == "rate limit exceeded" {
                code := http.StatusTooManyRequests
                c.JSON(code, gin.H{"success": false, "data": nil, "message": "rate limit exceeded", "responseCode": code})
                return
            }
            code := http.StatusInternalServerError
            c.JSON(code, gin.H{"success": false, "data": nil, "message": err.Error(), "responseCode": code})
            return
        }
        code := http.StatusOK
        c.JSON(code, gin.H{"success": true, "data": resp, "message": "", "responseCode": code})
	}
}
