package controller

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/core"
	errors2 "github.com/siahsang/blog/internal/errors"
	"github.com/siahsang/blog/internal/handler"
	"github.com/siahsang/blog/internal/validator"

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

func (u *ProfileController) Follow(ctx *gin.Context) {
	username := strings.TrimSpace(ctx.Param("username"))

	if username == "" {
		u.handler.HandleResponse(ctx, nil, &errors2.AppError{
			Code: http.StatusBadRequest,
			ErrorDetails: []*validator.FormErrorField{
				{
					ErrorField: "username",
					ErrorMsg:   "username is required",
				},
			},
		})
		return
	}

	authenticatedUser, err := auth.GetAuthenticatedUser(ctx)
	if err != nil {
		u.handler.HandleResponse(ctx, nil, &errors2.AppError{
			Code:         http.StatusUnauthorized,
			ErrorMessage: "Authentication required",
			ErrorStack:   err,
		})
		return
	}

	profile, err := u.core.FollowUser(ctx, *authenticatedUser, username)
	if err != nil {
		switch {
		case errors.Is(err, core.NoRecordFound):
			u.handler.HandleResponse(ctx, nil, &errors2.AppError{
				Code: http.StatusNotFound,
				ErrorDetails: []*validator.FormErrorField{
					{
						ErrorField: "body",
						ErrorMsg:   "User not found",
					},
				},
			})
			return
		case errors.Is(err, core.CannotFollowSelf):
			u.handler.HandleResponse(ctx, nil, &errors2.AppError{
				Code: http.StatusBadRequest,
				ErrorDetails: []*validator.FormErrorField{
					{
						ErrorField: "body",
						ErrorMsg:   "Cannot follow yourself",
					},
				},
			})
			return
		case errors.Is(err, core.UserIsAlreadyFollowed):
			u.handler.HandleResponse(ctx, nil, &errors2.AppError{
				Code: http.StatusBadRequest,
				ErrorDetails: []*validator.FormErrorField{
					{
						ErrorField: "body",
						ErrorMsg:   "User is already followed",
					},
				},
			})
			return
		default:
			u.handler.HandleResponse(ctx, nil, err)
			return
		}
	}

	u.handler.HandleResponse(ctx, map[string]any{"profile": profile}, nil)
}

func (u *ProfileController) Unfollow(ctx *gin.Context) {
	username := strings.TrimSpace(ctx.Param("username"))

	if username == "" {
		u.handler.HandleResponse(ctx, nil, &errors2.AppError{
			Code: http.StatusBadRequest,
			ErrorDetails: []*validator.FormErrorField{
				{
					ErrorField: "username",
					ErrorMsg:   "username is required",
				},
			},
		})
		return
	}

	authenticatedUser, err := auth.GetAuthenticatedUser(ctx)
	if err != nil {
		u.handler.HandleResponse(ctx, nil, &errors2.AppError{
			Code:         http.StatusUnauthorized,
			ErrorMessage: "Authentication required",
			ErrorStack:   err,
		})
		return
	}

	profile, err := u.core.UnfollowUser(ctx, *authenticatedUser, username)
	if err != nil {
		switch {
		case errors.Is(err, core.NoRecordFound):
			u.handler.HandleResponse(ctx, nil, &errors2.AppError{
				Code: http.StatusNotFound,
				ErrorDetails: []*validator.FormErrorField{
					{
						ErrorField: "body",
						ErrorMsg:   "User not found",
					},
				},
			})
			return
		case errors.Is(err, core.UserIsNotFollowed):
			u.handler.HandleResponse(ctx, nil, &errors2.AppError{
				Code: http.StatusBadRequest,
				ErrorDetails: []*validator.FormErrorField{
					{
						ErrorField: "body",
						ErrorMsg:   "User is not followed",
					},
				},
			})
			return
		default:
			u.handler.HandleResponse(ctx, nil, err)
			return
		}
	}

	u.handler.HandleResponse(ctx, map[string]any{"profile": profile}, nil)
}
