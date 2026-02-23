package router

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UIRouter struct {
	logger *slog.Logger
}

func NewUIRouter(logger *slog.Logger) *UIRouter {
	return &UIRouter{
		logger: logger,
	}
}

func (u *UIRouter) RegisterUIRouter(r *gin.Engine, baseURLPath string) {

	r.NoRoute(func(c *gin.Context) {
		u.logger.Error("No route found for request: %s", c.Request.URL.Path)
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource could not be found."})
	})
}
