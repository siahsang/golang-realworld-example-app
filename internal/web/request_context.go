package web

import "github.com/siahsang/blog/internal/auth"

type RequestContext struct {
	User  *auth.User
	Token string
}
