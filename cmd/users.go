package main

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/validator"
	"net/http"
	"strings"
	"time"
)

type envelope map[string]any

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	type registerUserPayload struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type RegisterUserRequest struct {
		registerUserPayload `json:"user"`
	}

	var registerUserRequest RegisterUserRequest

	if err := app.readJSON(w, r, &registerUserRequest); err != nil {
		app.badRequestResponse(w, r, &AppError{
			ErrorMessage: err.Error(),
			ErrorStack:   err,
		})
		return
	}

	user := &auth.User{
		Email:             registerUserRequest.Email,
		Username:          registerUserRequest.Username,
		PlaintextPassword: registerUserRequest.Password,
	}

	user.Email = strings.TrimSpace(user.Email)
	user.Username = strings.TrimSpace(user.Username)

	if err := user.SetPassword(registerUserRequest.Password); err != nil {
		app.internalErrorResponse(w, r, err)
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

	err := app.core.CreateNewUser(user)
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

	token, err := user.GenerateToken(time.Hour * 24 * 1)
	user.Token = token
	if err != nil {
		app.internalErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusAccepted, userResponse(user), nil); err != nil {
		app.internalErrorResponse(w, r, err)
	}

}

func (app *application) updateUser(w http.ResponseWriter, r *http.Request) {
	type updateUserPayload struct {
		Email string  `json:"email"`
		Bio   *string `json:"bio"`
		Image *string `json:"image"`
	}

	type UpdateUserRequest struct {
		updateUserPayload `json:"user"`
	}

	var updateUserRequest UpdateUserRequest

	if err := app.readJSON(w, r, &updateUserRequest); err != nil {
		app.badRequestResponse(w, r, &AppError{
			ErrorMessage: err.Error(),
			ErrorStack:   err,
		})
		return
	}

	authenticatedUser, _ := app.auth.GetAuthenticatedUser(r)

	if updateUserRequest.Bio != nil {
		trimmedBio := strings.TrimSpace(*updateUserRequest.Bio)
		authenticatedUser.Bio = &trimmedBio
	}

	if updateUserRequest.Image != nil {
		trimmedImage := strings.TrimSpace(*updateUserRequest.Image)
		authenticatedUser.Image = &trimmedImage
	}

	v := validator.New()
	checkEmail(v, updateUserRequest.Email)

	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{ErrorDetails: v.Errors})
		return
	}

	authenticatedUser.Email = strings.TrimSpace(updateUserRequest.Email)
	updateUser, err := app.core.UpdateUser(authenticatedUser)
	updateUser.Token = authenticatedUser.Token

	if err != nil {
		switch {
		case errors.Is(err, core.NoRecordFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.internalErrorResponse(w, r, err)
			return
		}
	}

	if err := app.writeJSON(w, http.StatusAccepted, userResponse(updateUser), nil); err != nil {
		app.internalErrorResponse(w, r, err)
	}

}

func (app *application) getUser(w http.ResponseWriter, r *http.Request) {
	authenticatedUser, _ := app.auth.GetAuthenticatedUser(r)

	if authenticatedUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, userResponse(authenticatedUser), nil); err != nil {
		app.internalErrorResponse(w, r, err)
	}
}

func (app *application) getProfile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	username := strings.TrimSpace(ps.ByName("username"))
	if username == "" {
		app.badRequestResponse(w, r, &AppError{
			ErrorMessage: "Username is required",
		})
		return
	}

	profile, err := app.core.GetProfile(username)
	if err != nil {
		switch {
		case errors.Is(err, core.NoRecordFound):
			app.badRequestResponse(w, r, &AppError{
				ErrorMessage: err.Error(),
				ErrorStack:   err,
			})
			return
		default:
			app.internalErrorResponse(w, r, err)
			return
		}
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"profile": profile}, nil); err != nil {
		app.internalErrorResponse(w, r, err)
	}
}

func (app *application) followUser(w http.ResponseWriter, r *http.Request) {
	parms := httprouter.ParamsFromContext(r.Context())
	authenticatedUser, _ := app.auth.GetAuthenticatedUser(r)

	followeeUsername := strings.TrimSpace(parms.ByName("followee"))
	v := validator.New()
	v.CheckNotBlank(followeeUsername, "followee", "must be provided")

	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{
			ErrorMessage: "Followee username must be provided",
		})
		return
	}

	profile, err := app.core.FollowUser(*authenticatedUser, followeeUsername)
	if err != nil {
		switch {
		case errors.Is(err, core.NoRecordFound):
			app.badRequestResponse(w, r, &AppError{
				ErrorMessage: err.Error(),
				ErrorStack:   err,
			})
			return
		case errors.Is(err, core.UserIsAlreadyFollowed):
			app.badRequestResponse(w, r, &AppError{
				ErrorMessage: err.Error(),
				ErrorStack:   err,
			})
			return
		default:
			app.internalErrorResponse(w, r, err)
			return
		}
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"profile": profile}, nil); err != nil {
		app.internalErrorResponse(w, r, err)
	}
}

func (app *application) unfollowUser(w http.ResponseWriter, r *http.Request) {
	parms := httprouter.ParamsFromContext(r.Context())
	authenticatedUser, _ := app.auth.GetAuthenticatedUser(r)

	followeeUsername := strings.TrimSpace(parms.ByName("followee"))
	v := validator.New()
	v.CheckNotBlank(followeeUsername, "followee", "must be provided")

	if !v.IsValid() {
		app.badRequestResponse(w, r, &AppError{
			ErrorMessage: "Followee username must be provided",
		})
		return
	}

	profile, err := app.core.UnfollowUser(*authenticatedUser, followeeUsername)
	if err != nil {
		switch {
		case errors.Is(err, core.UserIsNotFollowed):
			app.badRequestResponse(w, r, &AppError{
				ErrorMessage: err.Error(),
				ErrorStack:   err,
			})
			return
		default:
			app.internalErrorResponse(w, r, err)
			return
		}
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"profile": profile}, nil); err != nil {
		app.internalErrorResponse(w, r, err)
	}
}

func userResponse(user *auth.User) envelope {
	return envelope{"user": user}
}
