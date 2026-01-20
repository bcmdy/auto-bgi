package auth

import (
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("apiKey")
		if apiKey != "" && apiKey == User.Auth.APIKey {
			c.Next()
			return
		}
		QueryToken := c.Query("token")
		if QueryToken != "" {
			_, err := ParseToken(QueryToken)
			if err != nil {
				c.JSON(401, gin.H{"code": 401, "error": "未授权"})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		token := c.GetHeader("Authorization")
		_, err := ParseToken(token)
		if err != nil {
			c.JSON(401, gin.H{"code": 401, "error": "未授权"})
			c.Abort()
			return
		}
		c.Next()
	}
}
