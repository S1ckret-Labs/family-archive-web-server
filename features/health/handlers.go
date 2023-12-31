package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Get health status
// @Description Get the health status of the application
// @ID get-health
// @Produce json
// @Success 200 "Successful response"
// @Failure 400 "Bad Request"
// @Router /health [get]
func GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}
