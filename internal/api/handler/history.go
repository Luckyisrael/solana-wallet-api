package handler

import (
    "net/http"
    "fmt"
    "github.com/gin-gonic/gin"
    solanago "github.com/gagliardetto/solana-go"
    "github.com/Luckyisrael/solana-wallet-api/internal/service/history"
)

// GetHistory godoc
// @Summary      Get transaction history
// @Description  Returns recent transaction signatures for the given address
// @Tags         transactions
// @Produce      json
// @Param        address  path   string  true  "Wallet address"
// @Param        limit    query  int     false "Max items (default: 20)"
// @Param        before   query  string  false "Signature cursor"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /transactions/{address}/history [get]
func GetHistory(historyService *history.Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        addrStr := c.Param("address")
        if addrStr == "" {
            code := http.StatusBadRequest
            c.JSON(code, gin.H{"success": false, "data": nil, "message": "address is required", "responseCode": code})
            return
        }
        address, err := solanago.PublicKeyFromBase58(addrStr)
        if err != nil {
            code := http.StatusBadRequest
            c.JSON(code, gin.H{"success": false, "data": nil, "message": "invalid address", "responseCode": code})
            return
        }
        limit := 20
        if l := c.Query("limit"); l != "" {
            var n int
            _, _ = fmt.Sscanf(l, "%d", &n)
            if n > 0 {
                limit = n
            }
        }
        before := c.Query("before")
        resp, err := historyService.GetHistory(c.Request.Context(), address, limit, before)
        if err != nil {
            code := http.StatusInternalServerError
            c.JSON(code, gin.H{"success": false, "data": nil, "message": err.Error(), "responseCode": code})
            return
        }
        code := http.StatusOK
        c.JSON(code, gin.H{"success": true, "data": resp, "message": "", "responseCode": code})
    }
}
