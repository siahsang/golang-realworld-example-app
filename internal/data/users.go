package data

import (
	"context"
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type User struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
	ImageURL string `json:"imageURL"`
	password string `json:"-"`
}

type UserModel struct {
	DB  *sql.DB
	log *slog.Logger
}

func (userModel UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (email, token, username, bio, image)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
`
	args := []interface{}{user.Email, user.Token, user.Username, user.Bio, user.ImageURL}
	ctx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFunc()

	err := userModel.DB.QueryRowContext(ctx, query, args).Scan(&user.ID)

	return err
}

func (user User) SetPassword(plainTextPassword string) error {
	password, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)

	if err != nil {
		return err
	}

	user.password = password

	return nil

}
