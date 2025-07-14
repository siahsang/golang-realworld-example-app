package auth

import "github.com/golang-jwt/jwt/v5"

type User struct {
	ID                int64  `json:"-"`
	Email             string `json:"email"`
	Token             string `json:"token,omitempty"`
	Username          string `json:"username"`
	Password          []byte `json:"-"`
	PlaintextPassword string `json:"-"`
}

type UserClaim struct {
	Username string `json:"username"`
	Email    string `json:"email"`

	jwt.RegisteredClaims
}
