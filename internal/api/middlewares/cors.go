package middlewares

import (
	"os"

	"github.com/gin-gonic/gin"
)

func isAllowedOrigin(c *gin.Context) {
	appUrl := os.Getenv("APP_URL")
	allowList := map[string]bool{
		appUrl:                        true,
		"https://accounts.google.com": true,
	}

	if origin := c.Request.Header.Get("Origin"); allowList[origin] {
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	}

	return
}

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAllowedOrigin(c)
		//c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
