package router

import (
	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/controller"
)

type BlogAPIRouter struct {
	userController    *controller.UserController
	profileController *controller.ProfileController
	articleController *controller.ArticleController
	commentController *controller.CommentController
}

func NewBlogAPIRouter(
	userController *controller.UserController,
	profileController *controller.ProfileController,
	articleController *controller.ArticleController,
	commentController *controller.CommentController) *BlogAPIRouter {
	return &BlogAPIRouter{
		userController:    userController,
		profileController: profileController,
		articleController: articleController,
		commentController: commentController,
	}
}

func (b *BlogAPIRouter) RegisterOptionalAuthAPIRouter(group *gin.RouterGroup) {
	group.POST("/users", b.userController.CreateUser)
	group.POST("/users/login", b.userController.Login)

	group.GET("/profiles/:username", b.profileController.GetProfile)

	group.GET("/articles", b.articleController.GetArticles)
	group.GET("/articles/:slug", b.articleController.GetArticleBySlug)
	group.GET("/articles/:slug/comments", b.commentController.GetComments)
}

func (b *BlogAPIRouter) RegisterAuthRequiredAPIRouter(group *gin.RouterGroup) {
	group.POST("/profiles/:username/follow", b.profileController.Follow)
	group.DELETE("/profiles/:username/follow", b.profileController.Unfollow)

	group.GET("/articles/feed", b.articleController.Feed)
	group.POST("/articles", b.articleController.CreateArticle)
	group.PUT("/articles/:slug", b.articleController.UpdateArticle)
	group.DELETE("/articles/:slug", b.articleController.DeleteArticle)
	group.POST("/articles/:slug/comments", b.commentController.CreateComment)

}
