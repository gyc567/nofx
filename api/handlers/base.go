package handlers

import (
	"nofx/api/credits"
	"nofx/config"
	"nofx/email"
	"nofx/manager"
	creditsService "nofx/service/credits"

	"github.com/gin-gonic/gin"
)

// BaseHandler holds dependencies for API handlers
type BaseHandler struct {
	TraderManager *manager.TraderManager
	Database      *config.Database
	EmailClient   *email.ResendClient
	CreditService creditsService.Service
	CreditHandler *credits.Handler
}

// NewBaseHandler creates a new BaseHandler
func NewBaseHandler(
	traderManager *manager.TraderManager,
	database *config.Database,
	emailClient *email.ResendClient,
	creditService creditsService.Service,
	creditHandler *credits.Handler,
) *BaseHandler {
	return &BaseHandler{
		TraderManager: traderManager,
		Database:      database,
		EmailClient:   emailClient,
		CreditService: creditService,
		CreditHandler: creditHandler,
	}
}

// GetUserID extracts the user ID from the Gin context
func (h *BaseHandler) GetUserID(c *gin.Context) string {
	return c.GetString("user_id")
}
