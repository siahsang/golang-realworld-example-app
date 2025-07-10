package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/validator"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

var (
	ErrDuplicateEmail    = xerrors.Message("duplicate email")
	ErrDuplicateUsername = xerrors.Message("duplicate username")
	NoRecordFound        = xerrors.Message("record not found")
)

type User struct {
	ID                int64  `json:"-"`
	Email             string `json:"email"`
	Token             string `json:"token,omitempty"`
	Username          string `json:"username"`
	password          []byte `json:"-"`
	PlaintextPassword string `json:"-"`
}

type Claim struct {
	Username string `json:"username"`
	Email    string `json:"email"`

	jwt.RegisteredClaims
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := userModel.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return xerrors.New(err)
		}
	}

	return nil
}

func (userModel UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, email, username, password
		FROM users
		WHERE email = $1
	`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := userModel.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.password,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, NoRecordFound
		default:
			return nil, xerrors.New(err)
		}
	}

	return &user, nil
}

func (user *User) ValidateUser(v *validator.Validator) {
	// check email
	v.CheckNotBlank(user.Email, "email", "must be provided")
	v.CheckEmail(user.Email, "must be a valid email address")

	// check username
	v.CheckNotBlank(user.Username, "username", "must be provided")
	v.Check(len(user.Username) >= 5, "username", "must be at least 5 characters long")

	// check PlaintextPassword
	v.CheckNotBlank(user.PlaintextPassword, "Plaintext Password", "must be provided")
	v.Check(len(user.PlaintextPassword) >= 8, "Plaintext Password", "must be at least 8 characters long")

	// check password
	v.CheckNotBlank(string(user.password), "password", "must be provided")

}

func ValidateEmail(v *validator.Validator, email string) {
	v.CheckNotBlank(email, "email", "must be provided")
	v.CheckEmail(email, "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.CheckNotBlank(password, "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 characters long")
}

func (user *User) SetPassword(plainTextPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)

	if err != nil {
		return xerrors.New(err)
	}

	user.password = hashedPassword
	return nil
}

func (user *User) IsPasswordMatch(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(user.password, []byte(plainTextPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, xerrors.New(err)
	}

	return true, nil
}

func (user *User) GenerateToken(duration time.Duration) (string, error) {
	expireAt := time.Now().Add(duration)
	claim := Claim{
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	// todo: replace with the env variable
	signedString, err := token.SignedString([]byte("your-secret-key"))

	return signedString, xerrors.New(err)
}
