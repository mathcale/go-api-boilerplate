package middlewares

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/core/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/core/response"
	"github.com/mathcale/go-api-boilerplate/internal/core/webcontext"
)

type (
	RequireRoleMiddleware interface {
		Handler(next http.Handler) http.Handler
	}

	requireRoleMiddleware struct {
		acceptedRoles []string
		response      response.Response
	}
)

func NewRequireRoleMiddleware(
	response response.Response,
	acceptedRoles ...string,
) RequireRoleMiddleware {
	return &requireRoleMiddleware{
		acceptedRoles: acceptedRoles,
		response:      response,
	}
}

func (m *requireRoleMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roles, _ := webcontext.RolesFromContext(r.Context())

		if !hasAcceptedRole(roles, m.acceptedRoles) {
			forbidden := apperror.BE0004_FORBIDDEN

			m.response.RespondWithError(w, apperror.New(
				nil, "insufficient permissions", apperror.ForbiddenKind,
				apperror.MiddlewareOrigin, "require_role", &forbidden, nil,
			), nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func hasAcceptedRole(userRoles, acceptedRoles []string) bool {
	for _, accepted := range acceptedRoles {
		for _, owned := range userRoles {
			if owned == accepted {
				return true
			}
		}
	}

	return false
}
