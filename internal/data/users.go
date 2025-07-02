package data

import (
	"context"
	"database/sql"
	"github.com/siahsang/blog/internal/validator"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type User struct {
	ID                int64  `json:"id"`
	Email             string `json:"email"`
	Token             string `json:"token"`
	Username          string `json:"username"`
	Bio               string `json:"bio"`
	ImageURL          string `json:"imageURL"`
	password          []byte `json:"-"`
	PlaintextPassword string `json:"-"`
}

type UserModel struct {
	DB  *sql.DB
	log *slog.Logger
}

func (userModel UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id
`
	args := []interface{}{user.Username, user.Email, user.password}
	ctx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFunc()

	err := userModel.DB.QueryRowContext(ctx, query, args).Scan(&user.ID)

	return err
}

func (user User) ValidateUser(v *validator.Validator) {
	// check email
	v.CheckNotBlank(user.Email, "email", "must be provided")
	v.CheckEmail(user.Email, "must be a valid email address")

	// check username
	v.CheckNotBlank(user.Username, "username", "must be provided")
	v.Check(len(user.Username) >= 5, "username", "must be at least 5 characters long")

	// check PlaintextPassword
	v.CheckNotBlank(user.PlaintextPassword, "password", "must be provided")
	v.Check(len(user.PlaintextPassword) >= 8, "password", "must be at least 8 characters long")

	// check password
	v.CheckNotBlank(string(user.password), "password", "must be provided")

}

func (user User) SetPassword(plainTextPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)

	if err != nil {
		return err
	}

	user.password = hashedPassword
	return nil
}
