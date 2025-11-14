package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Luckyisrael/solana-wallet-api/internal/redis"
)

func RateLimitPerKey(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyID, exists := c.Get("api_key_id")
		if !exists {
			c.Next()
			return
		}

		limit, _ := c.Get("rate_limit")
		limitInt := limit.(int)

		key := fmt.Sprintf("ratelimit:key:%d", keyID)
		now := time.Now().Unix()
		window := 60 // 1 minute

		// Lua script for sliding window
		count, _ := redisClient.Eval(c, `
			local key = KEYS[1]
			local now = ARGV[1]
			local window = ARGV[2]
			local limit = ARGV[3]

			redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
			local current = redis.call('ZCARD', key)
			if current >= tonumber(limit) then
				return 1
			end
			redis.call('ZADD', key, now, now..':'..redis.call('INCR', key..':counter'))
			redis.call('EXPIRE', key, window)
			return 0
		`, []string{key}, now, window, limitInt).Result()

		if count.(int64) == 1 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"limit": limitInt,
				"period": "1 minute",
			})
			return
		}

		c.Next()
	}
}