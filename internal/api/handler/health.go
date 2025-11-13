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
    c.JSON(200, gin.H{
        "status": "ok",
        "time":   time.Now().UTC(),
    })
}
