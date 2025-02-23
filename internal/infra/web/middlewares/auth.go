package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type userIDCtx struct{}

type authMiddleware struct {
	logger logger.Logger
	jwt    jwt.JWTAuth
}

func NewAuthMiddleware(logger logger.Logger, jwt jwt.JWTAuth) MiddlewareHandler {
	return &authMiddleware{
		logger: logger,
		jwt:    jwt,
	}
}

func (m *authMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := m.parseToken(r.Header.Get("Authorization"))
		if err != nil {
			m.respondUnauthorized(w)
			return
		}

		m.logger.Debug("Valid token parsed", map[string]interface{}{
			"sub": token.Subject,
			"exp": token.ExpiresAt,
		})

		ctx := context.WithValue(r.Context(), userIDCtx{}, token.Subject)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (m *authMiddleware) parseToken(tokenStr string) (*jwt.Token, error) {
	errInvalidToken := errors.New("err_invalid_token")

	if tokenStr == "" {
		return nil, errInvalidToken
	}

	tokenSplit := strings.Split(tokenStr, " ")
	if len(tokenSplit) != 2 || tokenSplit[0] != "Bearer" {
		return nil, errInvalidToken
	}

	token, err := m.jwt.VerifyAccessToken(tokenSplit[1])
	if err != nil {
		m.logger.Error("Error while verifying token", err, nil)
		return nil, errInvalidToken
	}

	return token, nil
}

func (m *authMiddleware) respondUnauthorized(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"message":"Unauthorized"}`))
}
