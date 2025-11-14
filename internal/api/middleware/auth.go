package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/Luckyisrael/solana-wallet-api/internal/repo"
    "go.uber.org/zap"
)

type AuthMiddleware struct {
	apiKeyRepo *repo.APIKeyRepo
	logger     *zap.Logger
}

func NewAuthMiddleware(repo *repo.APIKeyRepo) *AuthMiddleware {
	logger, _ := zap.NewProduction()
	return &AuthMiddleware{apiKeyRepo: repo, logger: logger}
}

func (m *AuthMiddleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("X-API-Key")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			return
		}

		key, err := m.apiKeyRepo.ValidateKey(c.Request.Context(), auth)
		if err != nil || key == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or inactive API key"})
			return
		}

		// Attach to context
		c.Set("api_key_id", key.ID)
		c.Set("api_key_name", key.Name)
		c.Set("rate_limit", key.RateLimitPerMin)

		// Audit log
		m.logger.Info("API request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.Int("key_id", key.ID),
			zap.String("key_name", key.Name),
		)

		c.Next()
	}
}
