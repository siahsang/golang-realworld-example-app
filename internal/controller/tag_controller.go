package controller

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/handler"
	"github.com/siahsang/blog/internal/utils/functional"
	"github.com/siahsang/blog/models"
)

type TagController struct {
	core    *core.Core
	log     *slog.Logger
	handler *handler.Handler
}

func NewTagController(core *core.Core, log *slog.Logger) *TagController {
	return &TagController{
		core:    core,
		log:     log,
		handler: handler.NewHandler(log),
	}
}

func (t *TagController) GetTags(ctx *gin.Context) {
	tags, err := t.core.GetTagsList(ctx)
	if err != nil {
		t.handler.HandleResponse(ctx, nil, err)
		return
	}

	tagNames := functional.Map(tags, func(tag *models.Tag) string {
		return tag.Name
	})

	t.handler.HandleResponse(ctx, gin.H{"tags": tagNames}, nil)
}
