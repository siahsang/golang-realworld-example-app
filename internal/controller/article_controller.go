package controller

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/errors"
	"github.com/siahsang/blog/internal/filter"
	"github.com/siahsang/blog/internal/handler"
)

type ArticleController struct {
	core    *core.Core
	log     *slog.Logger
	auth    *auth.Auth
	handler *handler.Handler
}

func (c *ArticleController) getArticles(ctx *gin.Context) {
	v := validator.New()
	tagQ := ctx.Query("tag")
	authorQ := ctx.Query("author")
	favoritedQ := ctx.Query("favorited")
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit < 0 {
		c.handler.HandleResponse(ctx, nil, errors.AppError{
			ErrorMessage: "invalid limit",
		})
		return
	}

	offset, err2 := strconv.ParseInt(offsetStr, 10, 64)
	if err2 != nil || offset < 0 {
		c.handler.HandleResponse(ctx, nil, errors.AppError{
			ErrorMessage: "invalid offset",
		})
		return
	}

	filters := filter.NewFilter(limit, offset)

	articles, err := c.core.GetArticles(ctx, filters, tagQ, authorQ, favoritedQ)
	if err != nil {
		c.handler.HandleResponse(ctx, nil, err)
		return
	}

	user, _ := auth.GetAuthenticatedUser(ctx)
	response, err := prepareMultiArticleResponse(r, articles, c, user)
	if err != nil {
		c.internalErrorResponse(w, r, err)
		return
	}

	if err := c.writeJSON(w, http.StatusOK, response, nil); err != nil {
		c.internalErrorResponse(w, r, err)
		return
	}

}
