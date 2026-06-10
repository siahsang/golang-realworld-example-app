package server

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/router"
	"github.com/siahsang/blog/internal/utils/config"
)

func NewHttpServer(
	debug bool,
	uiRouter *router.UIRouter,
	uiConfig *UIConfig,
	blogAPIRouter *router.BlogAPIRouter,
	logger *slog.Logger,
	config config.Config,
	coreSys *core.Core,
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

	ginEngine.Use(func(context *gin.Context) {
		context.Header("Vary", "Authorization")

		autherization := context.GetHeader("Authorization")
		if autherization != "" {
			autherizationParts := strings.Split(autherization, " ")
			if len(autherizationParts) == 2 && autherizationParts[0] == "Token" {
				token := autherizationParts[1]
				claim, err := auth.ValidateToken(token, config.JWTSecret)
				if err == nil {
					user, err := coreSys.GetUserByEmail(context, claim.Email)
					if err == nil {
						user.Token = token
						auth.SetAuthenticatedUser(context, user)
					}
				}
			}
		}

		context.Next()
	})

	// handle 404 and static files
	uiRouter.RegisterUIRouter(ginEngine, uiConfig.APIBaseURL)

	// route must be available without logging in
	mustNotAuthGroupAPI := ginEngine.Group(uiConfig.APIBaseURL)
	blogAPIRouter.RegisterMustNotAuthAPIRouter(mustNotAuthGroupAPI)

	return ginEngine
}
