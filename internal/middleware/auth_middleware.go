package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func APIKeyAuthMiddleware(requiredKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyFromClient := c.GetHeader("X-API-KEY")
		if keyFromClient == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			return
		}

		if keyFromClient != requiredKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			return
		}

		c.Next()
	}
}
