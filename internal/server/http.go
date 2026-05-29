package server

import (
	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/router"
)

func NewHttpServer(
	debug bool,
	uiRouter *router.UIRouter,
	uiConfig *UIConfig,
	blogAPIRouter *router.BlogAPIRouter,
) *gin.Engine {

	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ginEngine := gin.Default()

	//static := ginEngine.Group(uiConfig.APIBaseURL)

	// handle 404 and static files
	uiRouter.RegisterUIRouter(ginEngine, uiConfig.APIBaseURL)

	// route must be available without logging in
	mustNotAuthGroupAPI := ginEngine.Group(uiConfig.APIBaseURL)
	blogAPIRouter.RegisterMustNotAuthAPIRouter(mustNotAuthGroupAPI)

	return ginEngine
}
