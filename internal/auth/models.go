package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/siahsang/blog/internal/utils/collectionutils"
)

type User struct {
	ID                int64   `json:"-"`
	Email             string  `json:"email"`
	Token             string  `json:"token,omitempty"`
	Username          string  `json:"username"`
	Password          []byte  `json:"-"`
	PlaintextPassword string  `json:"-"`
	Bio               *string `json:"bio"`
	Image             *string `json:"image"`
}

type UserClaim struct {
	Username string `json:"username"`
	Email    string `json:"email"`

	jwt.RegisteredClaims
}

type Auth struct {
	authenticatedUsers *collectionutils.SafeMap[string, *User]
}

func New() *Auth {
	return &Auth{}
}
