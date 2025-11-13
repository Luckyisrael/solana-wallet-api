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
			c.JSON(http.StatusBadRequest, gin.H{"error": "address is required"})
			return
		}

        resp, err := balanceService.GetBalance(c.Request.Context(), address)
        if err != nil {
            if err.Error() == "rate limit exceeded" {
                c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

		c.JSON(http.StatusOK, resp)
	}
}
