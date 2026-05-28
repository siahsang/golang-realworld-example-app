package controller

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/handler"
	"github.com/siahsang/blog/internal/schema"
)

type UserController struct {
	core *core.Core
	log  *slog.Logger
}

//func (u *UserController) Login(ctx *gin.Context) {
//
//	req := &schema.UserRegisterPayloadReq{}
//
//	if err := app.readJSON(w, r, &loginUserRequest); err != nil {
//		app.badRequestResponse(w, r, &AppError{
//			ErrorMessage: err.Error(),
//			ErrorStack:   err,
//		})
//		return
//	}
//
//	v := validator.New()
//
//	// check email
//	v.CheckNotBlank(loginUserRequest.Email, "email", "must be provided")
//	v.CheckEmail(loginUserRequest.Email, "must be a valid email address")
//
//	// check password
//	v.CheckNotBlank(loginUserRequest.Password, "password", "must be provided")
//
//	if !v.IsValid() {
//		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
//		return
//	}
//
//	user, err := app.core.GetUserByEmail(r.Context(), loginUserRequest.Email)
//	if err != nil {
//		switch {
//		case errors.Is(err, core.NoRecordFound):
//			app.badRequestResponse(w, r, &AppError{
//				ErrorMessage: "Invalid credentials",
//				ErrorStack:   err,
//			})
//			return
//		default:
//			app.internalErrorResponse(w, r, err)
//			return
//		}
//	}
//	match, err := user.IsPasswordMatch(loginUserRequest.Password)
//	if err != nil {
//		app.internalErrorResponse(w, r, err)
//	}
//	if !match {
//		app.badRequestResponse(w, r, &AppError{
//			ErrorMessage: "Invalid credentials",
//		})
//		return
//	}
//
//	token, err := user.GenerateToken(time.Hour*24*1, app.config.JWTSecret)
//	user.Token = token
//	if err != nil {
//		app.internalErrorResponse(w, r, err)
//		return
//	}
//
//	if err := app.writeJSON(w, http.StatusAccepted, userResponse(user), nil); err != nil {
//		app.internalErrorResponse(w, r, err)
//	}
//}

func (u *UserController) CreateUser(ctx *gin.Context) {

	req := &schema.UserRegisterReq{}

	if handler.BindAndCheck(ctx, req) {
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
		handler.HandleResponse(ctx, nil, err)
		return
	}

	v := validator.New()
	checkEmail(v, user.Email)

	// check username
	v.CheckNotBlank(user.Username, "username", "must be provided")
	v.Check(len(user.Username) >= 5, "username", "must be at least 5 characters long")

	// check PlaintextPassword
	v.CheckNotBlank(user.PlaintextPassword, "plaintext password", "must be provided")
	v.Check(len(user.PlaintextPassword) >= 8, "plaintext password", "must be at least 8 characters long")

	// check password
	v.CheckNotBlank(string(user.Password), "password", "must be provided")

	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
		return
	}

	err := u.core.CreateNewUser(r.Context(), user)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrDuplicateUsername):
			v.AddError("email", "Email address is already in use")
			app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
			return
		case errors.Is(err, core.ErrDuplicateEmail):
			v.AddError("username", "Username is already in use")
			app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
			return
		default:
			app.internalErrorResponse(w, r, err)
			return
		}
	}

	token, err := user.GenerateToken(time.Hour*24*1, app.config.JWTSecret)
	user.Token = token
	if err != nil {
		app.internalErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusAccepted, userResponse(user), nil); err != nil {
		app.internalErrorResponse(w, r, err)
	}

}
