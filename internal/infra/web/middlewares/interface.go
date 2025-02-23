package middlewares

import "net/http"

type MiddlewareHandler interface {
	Handler(next http.Handler) http.Handler
}
