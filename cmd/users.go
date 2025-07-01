package main

import (
	"github.com/siahsang/blog/internal/data"
	"github.com/siahsang/blog/internal/validator"
	"net/http"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Email    string `json:"email"`
		Password string `json:"token"`
		Username string `json:"username"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Email:    input.Email,
		Username: input.Username,
	}

	v := validator.New()

	err := user.SetPassword(input.Password)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User registered successfully"))
}
