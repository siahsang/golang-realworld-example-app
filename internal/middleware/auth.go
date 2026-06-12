package middleware

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/core"
	"github.com/siahsang/blog/internal/utils/config"
)

type AuthUserMiddleware struct {
	core   *core.Core
	config *config.Config
	log    *slog.Logger
}

// NewAuthUserMiddleware creates a new AuthUserMiddleware instance.
func NewAuthUserMiddleware(core *core.Core, config *config.Config, log *slog.Logger) *AuthUserMiddleware {
	return &AuthUserMiddleware{
		core:   core,
		config: config,
		log:    log,
	}
}

// OptionalAuthMiddleware creates a middleware that optionally authenticates requests.
// If a valid Token is provided in the Authorization header, the user is authenticated.
// Invalid or missing tokens are silently ignored (authentication is optional).
func (au *AuthUserMiddleware) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Header("Vary", "Authorization")

		authorization := context.GetHeader("Authorization")
		if authorization != "" {
			authorizationParts := strings.Split(authorization, " ")
			if len(authorizationParts) == 2 && authorizationParts[0] == "Token" {
				token := authorizationParts[1]
				claim, err := auth.ValidateToken(token, au.config.JWTSecret)
				if err == nil {
					user, err := au.core.GetUserByEmail(context, claim.Email)
					if err == nil && user != nil {
						user.Token = token
						auth.SetAuthenticatedUser(context, user)
					} else if err != nil {
						au.log.Warn("valid token but user not found", "email", claim.Email)
					}
				}
			}
		}

		context.Next()
	}
}
