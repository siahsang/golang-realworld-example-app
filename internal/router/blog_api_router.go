package router

import (
	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/controller"
)

type BlogAPIRouter struct {
	userController *controller.UserController
}




func NewBlogAPIRouter(userController *controller.UserController) *BlogAPIRouter {
	return &BlogAPIRouter{
		userController: userController,
	}
}

func (b *BlogAPIRouter) RegisterMustNotAuthAPIRouter(group *gin.RouterGroup) {
	group.POST("/users", b.userController.CreateUser)
	group.POST("/users/login", b.userController.Login)


}