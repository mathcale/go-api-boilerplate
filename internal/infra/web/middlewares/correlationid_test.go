package middlewares

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/infra/web/webcontext"
	"go-api-boilerplate/internal/pkg/logger"
)

type CorrelationIDMiddlewareTestSuite struct {
	suite.Suite
}

func TestCorrelationIDMiddleware(t *testing.T) {
	suite.Run(t, new(CorrelationIDMiddlewareTestSuite))
}

func (s *CorrelationIDMiddlewareTestSuite) serve(header *string) (string, string, map[string]any) {
	s.T().Helper()

	buf := &bytes.Buffer{}
	var ctxID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID, _ = webcontext.CorrelationIDFromContext(r.Context())
		logger.FromCtx(r.Context()).Info("handled")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if header != nil {
		req.Header.Set(correlationIDHeader, *header)
	}

	rec := httptest.NewRecorder()
	mw := NewCorrelationIDMiddleware(logger.NewWithWriter("info", buf))
	mw.Handler(next).ServeHTTP(rec, req)

	entry := map[string]any{}
	s.Require().NoError(json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry))

	return rec.Header().Get(correlationIDHeader), ctxID, entry
}

func (s *CorrelationIDMiddlewareTestSuite) TestHandler() {
	s.Run("should keep a valid incoming correlation id", func() {
		id := "abc.DEF_123-xyz"

		respID, ctxID, entry := s.serve(&id)

		s.Equal(id, respID)
		s.Equal(id, ctxID)
		s.Equal(id, entry["correlation_id"])
	})

	s.Run("should accept a 64 character correlation id", func() {
		id := strings.Repeat("a", 64)

		respID, _, _ := s.serve(&id)

		s.Equal(id, respID)
	})

	s.Run("should generate a uuid when the header is missing", func() {
		respID, ctxID, entry := s.serve(nil)

		s.Require().NoError(uuid.Validate(respID))
		s.Equal(respID, ctxID)
		s.Equal(respID, entry["correlation_id"])
	})

	invalid := map[string]string{
		"empty":             "",
		"invalid chars":     "bad id\"; drop",
		"longer than 64":    strings.Repeat("a", 65),
		"newline injection": "abc\\nINJECT",
	}

	for name, value := range invalid {
		s.Run("should replace an invalid correlation id: "+name, func() {
			respID, ctxID, entry := s.serve(&value)

			s.Require().NoError(uuid.Validate(respID))
			s.Equal(respID, ctxID)
			s.Equal(respID, entry["correlation_id"])
		})
	}
}
