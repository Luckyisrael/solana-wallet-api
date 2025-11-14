package handler

import (
    "time"
    "github.com/gin-gonic/gin"
)

// Health godoc
// @Summary      Health check
// @Description  Returns 200 if service is up
// @Tags         system
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func Health(c *gin.Context) {
    code := 200
    c.JSON(code, gin.H{"success": true, "data": gin.H{"status": "ok", "time": time.Now().UTC()}, "message": "", "responseCode": code})
}
