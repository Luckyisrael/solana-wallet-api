package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Luckyisrael/solana-wallet-api/internal/api/dto"
	"github.com/Luckyisrael/solana-wallet-api/internal/service/transfer"
)

// Transfer godoc
// @Summary      Build and sign a transfer
// @Description  Creates a signed SOL or SPL token transfer. Returns base64 tx.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        body  body  dto.TransferRequest  true  "Transfer request"
// @Success      200  {object}  dto.TransferResponse
// @Failure      400  {object}  map[string]string
// @Failure      429  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /transactions/transfer [post]
func Transfer(transferService *transfer.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.TransferRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		resp, err := transferService.Transfer(c.Request.Context(), &req)
		if err != nil {
			if err.Error() == "rate limit exceeded" {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}