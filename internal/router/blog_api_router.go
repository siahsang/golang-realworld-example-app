package router

import (
	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/controller"
)

type BlogAPIRouter struct {
	userController   *controller.UserController
	profileController *controller.ProfileController
}

func NewBlogAPIRouter(userController *controller.UserController, profileController *controller.ProfileController) *BlogAPIRouter {
	return &BlogAPIRouter{
		userController:    userController,
		profileController: profileController,
	}
}

func (b *BlogAPIRouter) RegisterMustNotAuthAPIRouter(group *gin.RouterGroup) {
	group.POST("/users", b.userController.CreateUser)
	group.POST("/users/login", b.userController.Login)
	group.GET("/profiles/:username", b.profileController.GetProfile)
}
