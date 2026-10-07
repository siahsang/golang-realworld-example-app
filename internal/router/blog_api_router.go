package router

import (
	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/controller"
)

type BlogAPIRouter struct {
	userController    *controller.UserController
	profileController *controller.ProfileController
	articleController *controller.ArticleController
}

func NewBlogAPIRouter(
	userController *controller.UserController,
	profileController *controller.ProfileController,
	articleController *controller.ArticleController) *BlogAPIRouter {
	return &BlogAPIRouter{
		userController:    userController,
		profileController: profileController,
		articleController: articleController,
	}
}

func (b *BlogAPIRouter) RegisterOptionalAuthAPIRouter(group *gin.RouterGroup) {
	group.POST("/users", b.userController.CreateUser)
	group.POST("/users/login", b.userController.Login)

	group.GET("/profiles/:username", b.profileController.GetProfile)

	group.GET("/articles", b.articleController.GetArticles)
	group.GET("/articles/:slug", b.articleController.GetArticleBySlug)
}

func (b *BlogAPIRouter) RegisterAuthRequiredAPIRouter(group *gin.RouterGroup) {
	group.POST("/profiles/:username/follow", b.profileController.Follow)
	group.DELETE("/profiles/:username/follow", b.profileController.Unfollow)

	group.GET("/articles/feed", b.articleController.Feed)
	group.POST("/articles/articles", b.articleController.CreateArticle)

}
