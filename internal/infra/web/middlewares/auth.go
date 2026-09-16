package middlewares

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/webcontext"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
)

type AuthMiddleware interface {
	Handler(next http.Handler) http.Handler
}

type authMiddleware struct {
	jwtAuth  jwt.JWTAuth
	response handlers.Response
}

func NewAuthMiddleware(jwtAuth jwt.JWTAuth, response handlers.Response) AuthMiddleware {
	return &authMiddleware{
		jwtAuth:  jwtAuth,
		response: response,
	}
}

func (m *authMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := extractBearerToken(r)
		if err != nil {
			m.response.RespondWithError(w, err, nil)
			return
		}

		verified, err := m.jwtAuth.VerifyAccessToken(token)
		if err != nil {
			unauthorized := apperror.BE0003_UNAUTHORIZED

			m.response.RespondWithError(w, apperror.New(
				err, "invalid access token", apperror.UnauthorizedKind,
				apperror.MiddlewareOrigin, "auth", &unauthorized, nil,
			), nil)
			return
		}

		userID, err := uuid.Parse(verified.Subject)
		if err != nil {
			unauthorized := apperror.BE0003_UNAUTHORIZED

			m.response.RespondWithError(w, apperror.New(
				err, "invalid subject claim", apperror.UnauthorizedKind,
				apperror.MiddlewareOrigin, "auth", &unauthorized, nil,
			), nil)
			return
		}

		ctx := webcontext.WithUserID(r.Context(), userID)
		ctx = webcontext.WithRoles(ctx, verified.ExtraClaims.Roles)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	unauthorized := apperror.BE0003_UNAUTHORIZED

	if header == "" {
		return "", apperror.New(
			nil, "missing authorization header", apperror.UnauthorizedKind,
			apperror.MiddlewareOrigin, "auth", &unauthorized, nil,
		)
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", apperror.New(
			nil, "malformed authorization header", apperror.UnauthorizedKind,
			apperror.MiddlewareOrigin, "auth", &unauthorized, nil,
		)
	}

	return parts[1], nil
}
