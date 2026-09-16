package middlewares

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/webcontext"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
)

// RequireRoleMiddleware guards a route so that only requests whose authenticated
// user carries at least one of the accepted roles are allowed through. It must be
// layered AFTER AuthMiddleware, which populates the roles in the request context.
type RequireRoleMiddleware interface {
	Handler(next http.Handler) http.Handler
}

type requireRoleMiddleware struct {
	acceptedRoles []string
	response      handlers.Response
}

func NewRequireRoleMiddleware(
	response handlers.Response,
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
