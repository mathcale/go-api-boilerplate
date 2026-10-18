package middlewares

import (
	"net/http"
	"regexp"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/webcontext"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

const correlationIDHeader = "X-Correlation-ID"

var validCorrelationID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

type correlationIDMiddleware struct {
	logger logger.Logger
}

func NewCorrelationIDMiddleware(l logger.Logger) MiddlewareHandler {
	return &correlationIDMiddleware{logger: l}
}

func (m *correlationIDMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get(correlationIDHeader)
		if !validCorrelationID.MatchString(correlationID) {
			correlationID = uuid.NewString()
		}

		ctx := webcontext.WithCorrelationID(r.Context(), correlationID)
		ctx = logger.WithCtx(ctx, m.logger.With("correlation_id", correlationID))
		w.Header().Set(correlationIDHeader, correlationID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
