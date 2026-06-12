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
	errors2 "github.com/siahsang/blog/internal/errors"
	"github.com/siahsang/blog/internal/handler"
	"github.com/siahsang/blog/internal/schema"
	"github.com/siahsang/blog/internal/utils/config"
	"github.com/siahsang/blog/internal/validator"
)

type UserController struct {
	core    *core.Core
	log     *slog.Logger
	config  *config.Config
	handler *handler.Handler
}

func NewUserController(core *core.Core, log *slog.Logger, config *config.Config) *UserController {
	return &UserController{
		core:    core,
		log:     log,
		config:  config,
		handler: handler.NewHandler(log),
	}
}

func (u *UserController) Login(ctx *gin.Context) {

	req := &schema.UserLoginPayloadReq{}

	if u.handler.BindAndCheck(ctx, req) {
		return
	}

	user, err := u.core.GetUserByEmail(ctx, req.Email)
	if err != nil {
		switch {
		case errors.Is(err, core.NoRecordFound):
			u.handler.HandleResponse(ctx, nil, &errors2.AppError{
				Code:       http.StatusBadRequest,
				ErrorStack: err,
			})
			return
		default:
			u.handler.HandleResponse(ctx, nil, err)
			return
		}
	}
	match, err := user.IsPasswordMatch(req.Password)
	if err != nil {
		u.handler.HandleResponse(ctx, nil, &errors2.AppError{
			Code:       http.StatusInternalServerError,
			ErrorStack: err,
		})
		return
	}
	if !match {
		u.handler.HandleResponse(ctx, nil, &errors2.AppError{
			ErrorMessage: "Invalid credentials",
			Code:         http.StatusUnauthorized,
			ErrorStack:   err,
		})

		return
	}

	token, err := user.GenerateToken(time.Hour*24*1, u.config.JWTSecret)
	user.Token = token
	if err != nil {
		u.handler.HandleResponse(ctx, nil, &errors2.AppError{
			Code:       http.StatusInternalServerError,
			ErrorStack: err,
		})
		return
	}

	u.handler.HandleResponse(ctx, map[string]any{"user": user}, nil)
}

func (u *UserController) CreateUser(ctx *gin.Context) {

	req := &schema.UserRegisterReq{}

	if u.handler.BindAndCheck(ctx, req) {
		return
	}

	user := &auth.User{
		Email:             req.Email,
		Username:          req.Username,
		PlaintextPassword: req.Password,
	}

	user.Email = strings.TrimSpace(user.Email)
	user.Username = strings.TrimSpace(user.Username)

	if err := user.SetPassword(req.Password); err != nil {
		u.handler.HandleResponse(ctx, nil, err)
		return
	}

	err := u.core.CreateNewUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrDuplicateUsername):
			u.handler.HandleResponse(ctx, nil, &errors2.AppError{
				Code:         http.StatusBadRequest,
				ErrorStack:   err,
				ErrorMessage: "Username is already in use",
			})
			return
		case errors.Is(err, core.ErrDuplicateEmail):
			u.handler.HandleResponse(ctx, nil, &errors2.AppError{
				Code:         http.StatusBadRequest,
				ErrorStack:   err,
				ErrorMessage: "Email address is already in use",
			})

			return
		default:
			u.handler.HandleResponse(ctx, nil, &errors2.AppError{
				Code:       http.StatusInternalServerError,
				ErrorStack: err,
			})
			return
		}
	}

	token, err := user.GenerateToken(time.Hour*24*1, u.config.JWTSecret)
	if err != nil {
		u.handler.HandleResponse(ctx, nil, &errors2.AppError{
			Code:       http.StatusInternalServerError,
			ErrorStack: err,
		})
		return
	}
	user.Token = token

	u.handler.HandleResponse(ctx, map[string]any{"user": user}, nil)
}

func (u *UserController) GetProfile(ctx *gin.Context) {
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

	// Extract and validate Authorization header
	user, err2 := auth.GetAuthenticatedUser(ctx)
	var currentUserID *int64 = nil
	if err2 == nil {
		currentUserID = &user.ID
	}
	// Get profile from core
	profile, err := u.core.GetProfileByUserName(ctx, username, currentUserID)
	if err != nil {
		// Check if error is NoRecordFound (check error message or wrapped error)
		if errors.Is(err, core.NoRecordFound) {
			u.handler.HandleResponse(ctx, nil, &errors2.AppError{
				Code: http.StatusNotFound,
				ErrorDetails: []*validator.FormErrorField{
					{
						ErrorField: "body",
						ErrorMsg:   "profile not found",
					},
				},
			})
			return
		}
		u.handler.HandleResponse(ctx, nil, err)
		return
	}

	u.handler.HandleResponse(ctx, map[string]any{"profile": profile}, nil)
}

func (u *UserController) Follow(ctx *gin.Context) {
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

func (u *UserController) Unfollow(ctx *gin.Context) {
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
