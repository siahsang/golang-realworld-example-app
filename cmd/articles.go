package main

import (
	"errors"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/filter"
	"github.com/siahsang/blog/internal/utils/collectionutils"
	"github.com/siahsang/blog/internal/utils/functional"
	"github.com/siahsang/blog/internal/validator"
	"github.com/siahsang/blog/models"
	"net/http"
	"strings"
	"time"
)

// todo: creating article and tags should be implemented in the same transaction
func (app *application) createArticle(w http.ResponseWriter, r *http.Request) {
	type input struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Body        string    `json:"body"`
		TagList     *[]string `json:"tagList"`
	}

	type CreateArticleRequest struct {
		input `json:"article"`
	}

	var requestPayload CreateArticleRequest

	if err := app.readJSON(w, r, &requestPayload); err != nil {
		app.badRequestResponse(w, r, &AppError{
			ErrorMessage: err.Error(),
			ErrorStack:   err,
		})
		return
	}

	v := validator.New()
	v.CheckNotBlank(requestPayload.Title, "title", "must be provided")
	v.CheckNotBlank(requestPayload.Description, "description", "must be provided")
	v.CheckNotBlank(requestPayload.Body, "body", "must be provided")

	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
		return
	}

	var createdTags []*models.Tag
	if requestPayload.TagList != nil && len(*requestPayload.TagList) > 0 {
		for _, tag := range *requestPayload.TagList {
			v.CheckNotBlank(tag, "tag", "must be provided")
		}
		if !v.IsValid() {
			app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
			return
		}

		var tagModels []*models.Tag
		for _, tag := range *requestPayload.TagList {
			tagModels = append(tagModels, &models.Tag{Name: strings.TrimSpace(tag)})
		}

		tags, err := app.core.CreateTag(tagModels)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrDuplicatedSlug):
				app.internalErrorResponse(w, r, err)
				return
			default:
				app.internalErrorResponse(w, r, err)
				return
			}
		}
		user, _ := app.auth.GetAuthenticatedUser(r)
		createdTags = tags
		slug := app.core.CreateSlug(requestPayload.Title)
		article, err := app.core.CreateArticle(&models.Article{
			Title:       requestPayload.Title,
			Description: requestPayload.Description,
			Body:        requestPayload.Body,
			Slug:        slug,
			AuthorID:    user.ID,
		}, createdTags)

		if err != nil {
			switch {
			case errors.Is(err, core.ErrDuplicatedSlug):
				v.AddError("slug", "Slug already exists")
				app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors, ErrorStack: err})
				return
			default:
				app.internalErrorResponse(w, r, err)
				return
			}
		}

		isFavorited, _ := app.core.IsFavouriteArticleByUser(article.ID, user)
		favouriteArticleCount, err := app.core.FavouriteArticleCount(article.ID)
		if err != nil {
			app.internalErrorResponse(w, r, err)
			return
		}
		if err := app.writeJSON(w, http.StatusAccepted, articleResponse(article, createdTags, isFavorited, favouriteArticleCount), nil); err != nil {
			app.internalErrorResponse(w, r, err)
		}
	}
}

func (app *application) getArticles(w http.ResponseWriter, r *http.Request) {
	v := validator.New()
	query := r.URL.Query()
	tagQ := app.readString(query, "tag", "")
	authorQ := app.readString(query, "author", "")
	favoritedQ := app.readString(query, "favorited", "")

	limit := app.readInt(query, "limit", 20, v)
	offset := app.readInt(query, "offset", 0, v)

	filters := filter.NewFilter(limit, offset)

	filter.ValidateFilters(filters, v)
	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
		return
	}

	articles, err := app.core.GetArticles(filters, tagQ, authorQ, favoritedQ)
	if err != nil {
		app.internalErrorResponse(w, r, err)
		return
	}

	user, _ := app.auth.GetAuthenticatedUser(r)
	response, err := prepareMultiArticleResponse(articles, app, user)
	if err != nil {
		app.internalErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, response, nil); err != nil {
		app.internalErrorResponse(w, r, err)
		return
	}

}

func articleResponse(article *models.Article, createdTags []*models.Tag, isFavorited bool, FavoritesCount int64) envelope {
	type output struct {
		Slug           string    `json:"slug"`
		Title          string    `json:"title"`
		Description    string    `json:"description"`
		Body           string    `json:"body"`
		TagList        []string  `json:"tagList"`
		CreatedAt      time.Time `json:"createdAt"`
		UpdatedAt      time.Time `json:"updatedAt"`
		Favorited      bool      `json:"favorited"`
		FavoritesCount int64     `json:"favoritesCount"`
	}

	tagsList := make([]string, len(createdTags))
	for i, tag := range createdTags {
		tagsList[i] = tag.Name
	}
	articleEnvelop := &output{
		Slug:           article.Slug,
		Title:          article.Title,
		Description:    article.Description,
		Body:           article.Body,
		TagList:        tagsList,
		CreatedAt:      article.CreatedAt,
		UpdatedAt:      article.UpdatedAt,
		Favorited:      isFavorited,
		FavoritesCount: FavoritesCount,
	}

	return envelope{
		"article": articleEnvelop,
	}
}

func prepareMultiArticleResponse(articles []*models.Article, app *application, currentLoginUser *auth.User) (envelope, error) {
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

	tagsByArticleId, err := app.core.GetTagsByArticleId(articlesIdList)
	if err != nil {
		return nil, err
	}

	favouriteArticleByArticleId, err := app.core.FavouriteArticleByArticleId(articlesIdList, currentLoginUser)
	if err != nil {
		return nil, xerrors.New(err)
	}
	favouriteCountByArticleId, err := app.core.FavouriteCountByArticleId(articlesIdList)
	userIdList := functional.Map(articles, func(article *models.Article) int64 {
		return article.AuthorID
	})
	listOfUser, err := app.core.GetUsersByIdList(userIdList)
	if err != nil {
		return nil, xerrors.New(err)
	}

	userByUserId := collectionutils.Associate(listOfUser, func(user *auth.User) (int64, *auth.User) {
		return user.ID, user
	})

	var followingUserList []*auth.User
	if currentLoginUser != nil {
		followingUserList, err = app.core.GetFollowingUserList(currentLoginUser.Username)
		if err != nil {
			return nil, xerrors.New(err)
		}
	}

	followingUserById := collectionutils.Associate(followingUserList, func(user *auth.User) (int64, bool) {
		return user.ID, true
	})

	var articlesEnvelop []ArticleEnvelope
	for _, article := range articles {
		tagsList := collectionutils.GetOrDefault(tagsByArticleId, article.AuthorID, []models.Tag{})
		tagNameList := functional.Map(tagsList, func(t models.Tag) string { return t.Name })
		isFavorited := favouriteArticleByArticleId[article.ID]
		favoritesCount := favouriteCountByArticleId[article.ID]
		a := ArticleEnvelope{
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
		articlesEnvelop = append(articlesEnvelop, a)
	}

	return envelope{
		"articles": articlesEnvelop,
	}, nil
}
