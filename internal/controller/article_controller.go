package controller

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/errors"
	"github.com/siahsang/blog/internal/filter"
	"github.com/siahsang/blog/internal/handler"
	"github.com/siahsang/blog/internal/utils/collectionutils"
	"github.com/siahsang/blog/internal/utils/config"
	"github.com/siahsang/blog/internal/utils/functional"
	"github.com/siahsang/blog/models"
)

type envelope map[string]any

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

	if limit < 0 {
		c.handler.HandleResponse(ctx, nil, errors.AppError{
			ErrorMessage: "invalid limit",
		})
		return
	}

	if offset < 0 {
		c.handler.HandleResponse(ctx, nil, errors.AppError{
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

func prepareMultiArticleResponse(ctx *gin.Context, articles []*models.Article, totalCount int64, core *core.Core, currentLoginUser *auth.User) (envelope, error) {
	return prepareArticleResponse(ctx, articles, totalCount, core, currentLoginUser, false)
}

func prepareSingleArticleResponse(ctx *gin.Context, article *models.Article, totalCount int64, core *core.Core, currentLoginUser *auth.User) (envelope, error) {
	return prepareArticleResponse(ctx, []*models.Article{article}, totalCount, core, currentLoginUser, true)
}

func prepareArticleResponse(ctx *gin.Context, articles []*models.Article, totalCount int64, core *core.Core, currentLoginUser *auth.User, singleResponse bool) (envelope, error) {
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

	favouriteArticleByArticleId, err := core.FavouriteArticleByArticleId(ctx, articlesIdList, currentLoginUser)
	if err != nil {
		return nil, xerrors.New(err)
	}
	favouriteCountByArticleId, err := core.FavouriteCountByArticleId(ctx, articlesIdList)
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
		return envelope{
			"article": articlesEnvelop[0],
		}, nil

	}

	return envelope{
		"articles":      articlesEnvelop,
		"articlesCount": totalCount,
	}, nil
}
