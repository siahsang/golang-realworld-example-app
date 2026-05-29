package server

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/router"
)

func NewHttpServer(
	debug bool,
	uiRouter *router.UIRouter,
	uiConfig *UIConfig,
	blogAPIRouter *router.BlogAPIRouter,
	logger *slog.Logger,
) *gin.Engine {

	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ginEngine := gin.New()

	ginEngine.Use(func(context *gin.Context) {
		start := time.Now()
		context.Next()

		logger.Info("request completed",
			"method", context.Request.Method,
			"path", context.Request.URL.Path,
			"status", context.Writer.Status(),
			"duration", time.Since(start),
		)
	})

	ginEngine.Use(gin.Recovery())

	//static := ginEngine.Group(uiConfig.APIBaseURL)

	// handle 404 and static files
	uiRouter.RegisterUIRouter(ginEngine, uiConfig.APIBaseURL)

	// route must be available without logging in
	mustNotAuthGroupAPI := ginEngine.Group(uiConfig.APIBaseURL)
	blogAPIRouter.RegisterMustNotAuthAPIRouter(mustNotAuthGroupAPI)

	return ginEngine
}
