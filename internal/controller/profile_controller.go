package controller

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/handler"
	"log/slog"
)

type ProfileController struct {
	core    *core.Core
	log     *slog.Logger
	auth    *auth.Auth
	handler *handler.Handler
}

func NewProfileController(core *core.Core, log *slog.Logger, auth *auth.Auth) *ProfileController {
	return &ProfileController{
		core:    core,
		log:     log,
		auth:    auth,
		handler: handler.NewHandler(log),
	}
}

func (p *ProfileController) GetProfile(ctx *gin.Context) {
	username := ctx.Param("username")
	if username == "" {
		ctx.JSON(http.StatusNotFound, map[string]any{
			"errors": map[string][]string{
				"body": {"User not found"},
			},
		})
		return
	}

	var followerID *int64
	authHeader := ctx.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Token" {
			token := parts[1]
			jwtSecret := os.Getenv("JWT_SECRET")
			claim, err := p.auth.Authenticate(token, jwtSecret)
			if err == nil {
				user, err := p.core.GetUserByEmail(ctx, claim.Email)
				if err == nil {
					followerID = &user.ID
				}
			}
		}
	}

	profile, err := p.core.GetProfileByUserName(ctx, username, followerID)
	if err != nil {
		if core.IsUserNotFound(err) {
			ctx.JSON(http.StatusNotFound, map[string]any{
				"errors": map[string][]string{
					"body": {"User not found"},
				},
			})
			return
		}
		p.handler.HandleResponse(ctx, nil, err)
		return
	}

	p.handler.HandleResponse(ctx, map[string]any{"profile": profile}, nil)
}
