package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/web"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

const (
	UserCtxKey  = "user_data"
	TokenCtxKey = "token"
)

var (
	NotAuthenticatesUser = xerrors.Message("Not authenticated user")
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
	claim := UserClaim{
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

func (auth *Auth) Authenticate(tokenString string) (*UserClaim, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &UserClaim{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, xerrors.New("unexpected signing method")
		}
		// todo: replace with the env variable
		return []byte("your-secret-key"), nil
	})

	if err != nil {
		return nil, xerrors.New(err)
	}

	if !parsedToken.Valid {
		return nil, xerrors.New("invalid token")
	}

	if claim, ok := parsedToken.Claims.(*UserClaim); ok {
		return claim, nil
	} else {
		return nil, xerrors.New("could not parse claims")
	}
}

func (auth *Auth) GetAuthenticatedUser(r *http.Request) (*User, error) {
	user, ok := web.GetValueFromContext[*User](r, UserCtxKey)
	if !ok {
		return nil, NotAuthenticatesUser
	}

	return user, nil
}

func (auth *Auth) SetAuthenticatedUser(r *http.Request, user *User) *http.Request {
	return web.AddValueToContext(r, UserCtxKey, user)
}

func (auth *Auth) CacheAuthenticatedUser(user *User) {
	auth.authenticatedUsers.Store(user.Username, user)
}

func (auth *Auth) IsUserAuthenticated(r *http.Request) bool {
	_, err := auth.GetAuthenticatedUser(r)
	return err == nil
}
