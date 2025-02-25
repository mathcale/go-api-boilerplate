package middlewares

import (
	"context"
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
	"github.com/rs/xid"
)

type correlationIDMiddleware struct {
	logger logger.Logger
}

func NewCorrelationIDMiddleware(logger logger.Logger) MiddlewareHandler {
	return &correlationIDMiddleware{
		logger: logger,
	}
}

func (m *correlationIDMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := xid.New().String()

		ctx := context.WithValue(r.Context(), "correlation_id", correlationID)
		r = r.WithContext(ctx)

		m.logger.SetGlobalValue("correlation_id", correlationID)
		w.Header().Add("X-Correlation-ID", correlationID)

		next.ServeHTTP(w, r)
	})
}
