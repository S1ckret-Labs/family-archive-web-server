package helpers

import (
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		// TODO: If c.Writer.Status() >= 500, hide error messages for production
		c.JSON(-1, gin.H{
			"errors": c.Errors.Errors(),
		})
	}
}
