package server

import (
	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/router"
)

func NewHttpServer(debug bool,
	uiRouter *router.UIRouter,
	uiConfig *UIConfig) *gin.Engine {

	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ginEngine := gin.New()
	ginEngine.Use(gin.Logger())

	//static := ginEngine.Group(uiConfig.APIBaseURL)

	// handle 404 and static files
	uiRouter.RegisterUIRouter(ginEngine, uiConfig.APIBaseURL)

	// handle must not be authenticated routes

	return ginEngine
}
