package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleHealth checks the service health
func (h *BaseHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   c.Request.Context().Value("time"),
	})
}
