package health

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status string `json:"status"`
}

// RegisterHealthEndpoint registers the /health endpoint with the gin engine
func RegisterHealthEndpoint(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthStatus{Status: "UP"})
	})
}
