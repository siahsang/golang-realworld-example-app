package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mdobak/go-xerrors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func (user *User) SetPassword(plainTextPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)

	if err != nil {
		return xerrors.New(err)
	}

	user.Password = hashedPassword
	return nil
}

func (user *User) IsPasswordMatch(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(user.Password, []byte(plainTextPassword))
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
