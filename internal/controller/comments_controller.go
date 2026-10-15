package controller

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/handler"
	"github.com/siahsang/blog/internal/schema"
	"github.com/siahsang/blog/internal/utils/config"
	"github.com/siahsang/blog/internal/utils/databaseutils"
	"github.com/siahsang/blog/internal/utils/httperror"
	"github.com/siahsang/blog/models"
)

type CommentController struct {
	core    *core.Core
	log     *slog.Logger
	auth    *auth.Auth
	handler *handler.Handler
}

func NewCommentController(core *core.Core, logger *slog.Logger, cfg *config.Config) *CommentController {
	return &CommentController{
		core:    core,
		log:     logger,
		auth:    auth.New(cfg),
		handler: handler.NewHandler(logger),
	}
}

func (c *CommentController) CreateComment(ctx *gin.Context) {
	var createCommentRequest schema.CreateCommentRequest

	if c.handler.BindAndCheck(ctx, &createCommentRequest) {
		return
	}

	user, err := auth.GetAuthenticatedUser(ctx)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, &httperror.Error{
			Code:         http.StatusUnauthorized,
			ErrorMessage: "authentication required",
			ErrorStack:   err,
		})
		return
	}

	article, err := c.core.GetArticleBySlug(ctx, ctx.Param("slug"))
	if err != nil {
		if errors.Is(err, databaseutils.ErrNoRowsFound) {
			c.handler.HandleResponse(ctx, nil, &httperror.Error{
				Code:         http.StatusNotFound,
				ErrorMessage: "article not found",
			})
			return
		}
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	now := time.Now()
	newComment, err := c.core.CreateComment(ctx, &models.Comment{
		Body:      createCommentRequest.Comment.Body,
		CreatedAt: now,
		UpdatedAt: now,
		ArticleID: article.ID,
		AuthorID:  user.ID,
	})
	if err != nil {
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"comment": gin.H{
			"id":        newComment.ID,
			"body":      newComment.Body,
			"createdAt": newComment.CreatedAt,
			"updatedAt": newComment.UpdatedAt,
			"author": gin.H{
				"username":  user.Username,
				"bio":       user.Bio,
				"image":     user.Image,
				"following": false,
			},
		},
	})
}

func (c *CommentController) GetComments(ctx *gin.Context) {
	slug := ctx.Param("slug")
	if strings.TrimSpace(slug) == "" {
		c.handler.HandleResponse(ctx, nil, &httperror.Error{
			Code:         http.StatusBadRequest,
			ErrorMessage: "slug must be provided",
		})
		return
	}

	commentsBySlug, err := c.core.GetCommentsBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, databaseutils.ErrNoRowsFound) {
			c.handler.HandleResponse(ctx, nil, &httperror.Error{
				Code:         http.StatusNotFound,
				ErrorMessage: "article not found",
			})
			return
		}
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	var currentUserID *int64
	if user, err := auth.GetAuthenticatedUser(ctx); err == nil {
		currentUserID = &user.ID
	}

	comments := make([]gin.H, 0, len(commentsBySlug))
	profiles := make(map[int64]*models.Profile)
	for _, comment := range commentsBySlug {
		profile, ok := profiles[comment.AuthorID]
		if !ok {
			author, err := c.core.GetUsersById(ctx, comment.AuthorID)
			if err != nil {
				c.handler.HandleResponse(ctx, nil, err)
				return
			}
			profile, err = c.core.GetProfileByUserName(ctx, author.Username, currentUserID)
			if err != nil {
				c.handler.HandleResponse(ctx, nil, err)
				return
			}
			profiles[comment.AuthorID] = profile
		}
		comments = append(comments, gin.H{
			"id":        comment.ID,
			"body":      comment.Body,
			"createdAt": comment.CreatedAt,
			"updatedAt": comment.UpdatedAt,
			"author": gin.H{
				"username":  profile.Username,
				"bio":       profile.Bio,
				"image":     profile.Image,
				"following": profile.Following,
			},
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"comments": comments})
}
