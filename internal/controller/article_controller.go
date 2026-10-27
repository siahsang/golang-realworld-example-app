package controller

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/filter"
	"github.com/siahsang/blog/internal/handler"
	"github.com/siahsang/blog/internal/schema"
	"github.com/siahsang/blog/internal/utils/collectionutils"
	"github.com/siahsang/blog/internal/utils/config"
	"github.com/siahsang/blog/internal/utils/databaseutils"
	"github.com/siahsang/blog/internal/utils/functional"
	"github.com/siahsang/blog/internal/utils/httperror"
	"github.com/siahsang/blog/internal/utils/stringutils"
	"github.com/siahsang/blog/models"
)

type ArticleListQuery struct {
	Tag       string `form:"tag"`
	Author    string `form:"author"`
	Favorited string `form:"favorited"`
	Limit     int64  `form:"limit"`
	Offset    int64  `form:"offset"`
}

type ArticleController struct {
	core    *core.Core
	log     *slog.Logger
	auth    *auth.Auth
	handler *handler.Handler
}

func NewArticleController(core *core.Core, logger *slog.Logger, cfg *config.Config) *ArticleController {
	return &ArticleController{
		core:    core,
		log:     logger,
		auth:    auth.New(cfg),
		handler: handler.NewHandler(logger),
	}
}

func (c *ArticleController) GetArticles(ctx *gin.Context) {
	query := &ArticleListQuery{
		Limit:  20,
		Offset: 0,
	}

	if c.handler.BindAndCheck(ctx, query) {
		return
	}

	tag := query.Tag
	author := query.Author
	favorited := query.Favorited
	limit := query.Limit
	offset := query.Offset

	if limit <= 0 {
		c.handler.HandleResponse(ctx, nil, &httperror.Error{
			Code:         http.StatusBadRequest,
			ErrorMessage: "invalid limit",
		})
		return
	}

	if offset < 0 {
		c.handler.HandleResponse(ctx, nil, &httperror.Error{
			Code:         http.StatusBadRequest,
			ErrorMessage: "invalid offset",
		})
		return
	}

	filters := filter.NewFilter(limit, offset)

	articles, totalCount, err := c.core.GetArticles(ctx, filters, tag, author, favorited)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	user, _ := auth.GetAuthenticatedUser(ctx)
	response, err := prepareMultiArticleResponse(ctx, articles, totalCount, c.core, user)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	c.handler.HandleResponse(ctx, response, nil)
}

func (c *ArticleController) Feed(ctx *gin.Context) {
	user, _ := auth.GetAuthenticatedUser(ctx)

	query := &ArticleListQuery{
		Limit:  20,
		Offset: 0,
	}

	if c.handler.BindAndCheck(ctx, query) {
		return
	}

	limit := query.Limit
	offset := query.Offset

	if limit <= 0 {
		c.handler.HandleResponse(ctx, nil, &httperror.Error{
			Code:         http.StatusBadRequest,
			ErrorMessage: "invalid limit",
		})
		return
	}

	if offset < 0 {
		c.handler.HandleResponse(ctx, nil, &httperror.Error{
			Code:         http.StatusBadRequest,
			ErrorMessage: "invalid offset",
		})
		return
	}

	articles, totalCount, err := c.core.FeedArticle(ctx, user.Username, limit, offset)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	response, err := prepareMultiArticleResponse(ctx, articles, totalCount, c.core, user)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	c.handler.HandleResponse(ctx, response, nil)
}

func (c *ArticleController) GetArticleBySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")
	user, _ := auth.GetAuthenticatedUser(ctx)

	article, err := c.core.GetArticleBySlug(ctx, slug)

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

	response, err := prepareSingleArticleResponse(ctx, article, c.core, user)

	if err != nil {
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	c.handler.HandleResponse(ctx, response, nil)
}

func (c *ArticleController) CreateArticle(ctx *gin.Context) {
	var request schema.ArticlePayload
	if c.handler.BindAndCheck(ctx, &request) {
		return
	}

	request.Title = strings.TrimSpace(request.Title)
	request.Description = strings.TrimSpace(request.Description)
	request.Body = strings.TrimSpace(request.Body)

	user, err := auth.GetAuthenticatedUser(ctx)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, &httperror.Error{
			Code:         http.StatusUnauthorized,
			ErrorMessage: "authentication required",
			ErrorStack:   err,
		})
		return
	}

	tagModels := make([]*models.Tag, 0)
	if request.TagList != nil {
		seen := make(map[string]struct{}, len(*request.TagList))
		for _, value := range *request.TagList {
			name := strings.TrimSpace(value)
			if name == "" {
				continue
			}
			if _, exists := seen[name]; exists {
				continue
			}

			seen[name] = struct{}{}
			tagModels = append(tagModels, &models.Tag{Name: name})
		}
	}

	var article *models.Article
	slug := c.core.CreateSlug(request.Title)
	for {
		article, err = c.core.CreateArticle(ctx, &models.Article{
			Title:       request.Title,
			Description: request.Description,
			Body:        request.Body,
			Slug:        slug,
			AuthorID:    user.ID,
		}, tagModels)

		if err != nil {
			if errors.Is(err, core.ErrDuplicatedSlug) {
				slug = slug + "-" + stringutils.RandString(5)
				continue
			}

			c.handler.HandleResponse(ctx, nil, err)
			return
		}

		break
	}

	response, err := prepareSingleArticleResponse(ctx, article, c.core, user)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

func (c *ArticleController) UpdateArticle(ctx *gin.Context) {
	var updateArticleRequest schema.ArticleUpdatePayloadReq
	if c.handler.BindAndCheck(ctx, &updateArticleRequest) {
		return
	}

	slug := ctx.Param("slug")
	existingArticleBySlug, err := c.core.GetArticleBySlug(ctx, slug)

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

	user, err := auth.GetAuthenticatedUser(ctx)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, &httperror.Error{
			Code:         http.StatusUnauthorized,
			ErrorMessage: "authentication required",
			ErrorStack:   err,
		})
		return
	}

	if existingArticleBySlug.AuthorID != user.ID {
		c.handler.HandleResponse(ctx, nil, &httperror.Error{
			Code:         http.StatusForbidden,
			ErrorMessage: "you are not the author of this article",
		})
		return
	}

	isTitleChanged := false
	if updateArticleRequest.Title != nil && strings.TrimSpace(*updateArticleRequest.Title) != "" {
		newTrimTitle := strings.TrimSpace(*updateArticleRequest.Title)
		isTitleChanged = newTrimTitle != existingArticleBySlug.Title
		existingArticleBySlug.Title = newTrimTitle
	}

	if updateArticleRequest.Description != nil {
		trimSpace := strings.TrimSpace(*updateArticleRequest.Description)
		existingArticleBySlug.Description = trimSpace
	}
	if updateArticleRequest.Body != nil {
		trimSpace := strings.TrimSpace(*updateArticleRequest.Body)
		existingArticleBySlug.Body = trimSpace
	}

	newSlug := existingArticleBySlug.Slug
	if isTitleChanged {
		newSlug = c.core.CreateSlug(existingArticleBySlug.Title)
		existingArticleBySlug.Slug = newSlug
	}
	var updatedArticle *models.Article
	for {
		updatedArticle, err = c.core.UpdateArticle(ctx, existingArticleBySlug)
		if err != nil {
			if errors.Is(err, core.ErrDuplicatedSlug) {
				newSlug = newSlug + "-" + stringutils.RandString(5)
				existingArticleBySlug.Slug = newSlug
				continue
			}
			c.handler.HandleResponse(ctx, nil, err)
			return
		}
		break
	}

	response, err := prepareSingleArticleResponse(ctx, updatedArticle, c.core, user)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *ArticleController) DeleteArticle(ctx *gin.Context) {
	slug := ctx.Param("slug")
	existingArticleBySlug, err := c.core.GetArticleBySlug(ctx, slug)

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

	user, err := auth.GetAuthenticatedUser(ctx)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, &httperror.Error{
			Code:         http.StatusUnauthorized,
			ErrorMessage: "authentication required",
			ErrorStack:   err,
		})
		return
	}

	if existingArticleBySlug.AuthorID != user.ID {
		c.handler.HandleResponse(ctx, nil, &httperror.Error{
			Code:         http.StatusForbidden,
			ErrorMessage: "you are not the author of this article",
		})
		return
	}

	err = c.core.DeleteArticle(ctx, existingArticleBySlug.Slug)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func (c *ArticleController) FavoriteArticle(ctx *gin.Context) {
	slug := strings.TrimSpace(ctx.Param("slug"))
	if slug == "" {
		c.handler.HandleResponse(ctx, nil, &httperror.Error{
			Code:         http.StatusBadRequest,
			ErrorMessage: "slug must be provided",
		})
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

	article, err := c.core.FavoriteArticle(ctx, slug, user)
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

	response, err := prepareSingleArticleResponse(ctx, article, c.core, user)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	c.handler.HandleResponse(ctx, response, nil)
}

func (c *ArticleController) UnfavoriteArticle(ctx *gin.Context) {
	slug := strings.TrimSpace(ctx.Param("slug"))
	if slug == "" {
		c.handler.HandleResponse(ctx, nil, &httperror.Error{
			Code:         http.StatusBadRequest,
			ErrorMessage: "slug must be provided",
		})
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

	article, err := c.core.UnFavoriteArticle(ctx, slug, user)
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

	response, err := prepareSingleArticleResponse(ctx, article, c.core, user)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	c.handler.HandleResponse(ctx, response, nil)
}

func prepareMultiArticleResponse(ctx *gin.Context, articles []*models.Article, totalCount int64, core *core.Core, currentLoginUser *auth.User) (gin.H, error) {
	return prepareArticleResponse(ctx, articles, totalCount, core, currentLoginUser, false)
}

func prepareSingleArticleResponse(ctx *gin.Context, article *models.Article, core *core.Core, currentLoginUser *auth.User) (gin.H, error) {
	return prepareArticleResponse(ctx, []*models.Article{article}, 1, core, currentLoginUser, true)
}

func prepareArticleResponse(ctx *gin.Context, articles []*models.Article, totalCount int64, core *core.Core, currentLoginUser *auth.User, singleResponse bool) (gin.H, error) {
	type AuthorEnvelop struct {
		Username  string  `json:"username"`
		Bio       *string `json:"bio"`
		Image     *string `json:"image"`
		Following bool    `json:"following"`
	}

	type ArticleEnvelope struct {
		Slug           string        `json:"slug"`
		Title          string        `json:"title"`
		Description    string        `json:"description"`
		Body           *string       `json:"body,omitempty"`
		TagList        []string      `json:"tagList"`
		CreatedAt      time.Time     `json:"createdAt"`
		UpdatedAt      time.Time     `json:"updatedAt"`
		Favorited      bool          `json:"favorited"`
		FavoritesCount int64         `json:"favoritesCount"`
		Author         AuthorEnvelop `json:"author"`
	}

	articlesIdList := functional.Map(articles, func(a *models.Article) int64 {
		return a.ID
	})

	tagsByArticleId, err := core.GetTagsByArticleId(ctx, articlesIdList)
	if err != nil {
		return nil, err
	}

	favouriteArticleByArticleId, err := core.FavouriteArticleByUser(ctx, articlesIdList, currentLoginUser)
	if err != nil {
		return nil, xerrors.New(err)
	}
	favouriteCountByArticleId, err := core.FavouriteCountByArticleId(ctx, articlesIdList)
	if err != nil {
		return nil, xerrors.New(err)
	}

	userIdList := functional.Map(articles, func(article *models.Article) int64 {
		return article.AuthorID
	})
	listOfUser, err := core.GetUsersByIdList(ctx, userIdList)
	if err != nil {
		return nil, xerrors.New(err)
	}

	userByUserId := collectionutils.Associate(listOfUser, func(user *auth.User) (int64, *auth.User) {
		return user.ID, user
	})

	var followingUserList []*auth.User
	if currentLoginUser != nil {
		followingUserList, err = core.GetFollowingUserList(ctx, currentLoginUser.Username)
		if err != nil {
			return nil, xerrors.New(err)
		}
	}

	followingUserById := collectionutils.Associate(followingUserList, func(user *auth.User) (int64, bool) {
		return user.ID, true
	})

	articlesEnvelop := []ArticleEnvelope{}
	for _, article := range articles {
		tagsList := collectionutils.GetOrDefault(tagsByArticleId, article.ID, []models.Tag{})
		tagNameList := functional.Map(tagsList, func(t models.Tag) string { return t.Name })
		isFavorited := favouriteArticleByArticleId[article.ID]
		favoritesCount := favouriteCountByArticleId[article.ID]
		articleEnvelope := ArticleEnvelope{
			Slug:           article.Slug,
			Title:          article.Title,
			Description:    article.Description,
			TagList:        tagNameList,
			CreatedAt:      article.CreatedAt,
			UpdatedAt:      article.UpdatedAt,
			Favorited:      isFavorited,
			FavoritesCount: favoritesCount,
			Author: AuthorEnvelop{
				Username:  userByUserId[article.AuthorID].Username,
				Bio:       userByUserId[article.AuthorID].Bio,
				Image:     userByUserId[article.AuthorID].Image,
				Following: collectionutils.GetOrDefault(followingUserById, article.AuthorID, false),
			},
		}

		if singleResponse {
			articleEnvelope.Body = &article.Body
		}
		articlesEnvelop = append(articlesEnvelop, articleEnvelope)
	}

	if singleResponse {
		return gin.H{
			"article": articlesEnvelop[0],
		}, nil

	}

	return gin.H{
		"articles":      articlesEnvelop,
		"articlesCount": totalCount,
	}, nil
}
