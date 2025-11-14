package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Luckyisrael/solana-wallet-api/internal/api/dto"
	"github.com/Luckyisrael/solana-wallet-api/internal/service/broadcast"
)

// Broadcast godoc
// @Summary      Broadcast signed transaction
// @Description  Submits a base64-encoded signed tx. Idempotent.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        body  body  dto.BroadcastRequest  true  "Broadcast request"
// @Success      200  {object}  dto.BroadcastResponse
// @Failure      400  {object}  map[string]string
// @Failure      429  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /transactions/broadcast [post]
func Broadcast(broadcastService *broadcast.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.BroadcastRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            code := http.StatusBadRequest
            c.JSON(code, gin.H{"success": false, "data": nil, "message": "invalid request", "responseCode": code})
            return
        }

		resp, err := broadcastService.Broadcast(c.Request.Context(), &req)
        if err != nil {
            if err.Error() == "rate limit exceeded" {
                code := http.StatusTooManyRequests
                c.JSON(code, gin.H{"success": false, "data": nil, "message": err.Error(), "responseCode": code})
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
