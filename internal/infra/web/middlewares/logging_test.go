package middlewares

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/pkg/logger"
)

type LoggingMiddlewareTestSuite struct {
	suite.Suite
	buf     *bytes.Buffer
	handler func(next http.Handler) http.Handler
}

func TestLoggingMiddleware(t *testing.T) {
	suite.Run(t, new(LoggingMiddlewareTestSuite))
}

func (s *LoggingMiddlewareTestSuite) SetupTest() {
	s.buf = &bytes.Buffer{}
	log := logger.NewWithWriter("info", s.buf)

	chain := func(next http.Handler) http.Handler {
		return NewCorrelationIDMiddleware(log).Handler(NewLoggingMiddleware(log).Handler(next))
	}
	s.handler = chain
}

func (s *LoggingMiddlewareTestSuite) entries() []map[string]any {
	s.T().Helper()

	var out []map[string]any

	for _, line := range bytes.Split(bytes.TrimSpace(s.buf.Bytes()), []byte("\n")) {
		entry := map[string]any{}
		s.Require().NoError(json.Unmarshal(line, &entry))
		out = append(out, entry)
	}

	return out
}

func (s *LoggingMiddlewareTestSuite) do(next http.HandlerFunc, id string) *httptest.ResponseRecorder {
	s.T().Helper()

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set(correlationIDHeader, id)
	rec := httptest.NewRecorder()

	s.handler(next).ServeHTTP(rec, req)

	return rec
}

func (s *LoggingMiddlewareTestSuite) TestHandler() {
	s.Run("should log the request with its own correlation id", func() {
		ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		s.do(ok, "first")
		s.do(ok, "second")

		entries := s.entries()
		s.Require().Len(entries, 2)
		s.Equal("first", entries[0]["correlation_id"])
		s.Equal("second", entries[1]["correlation_id"])
		s.EqualValues(http.StatusTeapot, entries[0]["status_code"])
		s.Equal("Incoming request", entries[0]["message"])
	})

	s.Run("should not leak fields between sequential requests", func() {
		s.buf.Reset()
		ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		s.do(ok, "one")
		s.do(ok, "two")

		for i, entry := range s.entries() {
			s.Equal([]string{"one", "two"}[i], entry["correlation_id"])
		}
		s.NotContains(s.buf.String(), `"correlation_id":"one","correlation_id"`)
	})

	s.Run("should log with the base logger when none is in the context", func() {
		s.buf.Reset()
		log := logger.NewWithWriter("info", s.buf)
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		NewLoggingMiddleware(log).Handler(next).ServeHTTP(httptest.NewRecorder(), req)

		entries := s.entries()
		s.Require().Len(entries, 1)
		s.Equal("Incoming request", entries[0]["message"])
	})

	s.Run("should respond 500 with a JSON error when a panic happens before writing", func() {
		s.buf.Reset()
		boom := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("kaboom")
		})

		rec := s.do(boom, "p1")

		s.Equal(http.StatusInternalServerError, rec.Code)
		s.Equal("application/json", rec.Header().Get("Content-Type"))
		s.JSONEq(`{"message":"internal server error"}`, rec.Body.String())

		entries := s.entries()
		s.Require().Len(entries, 2)
		s.Equal("error", entries[0]["level"])
		s.Equal("panic: kaboom", entries[0]["error"])
		s.Equal("p1", entries[0]["correlation_id"])
		s.Contains(entries[0]["stack"], "logging_test.go")
		s.EqualValues(http.StatusInternalServerError, entries[1]["status_code"])
	})

	s.Run("should not write again when a panic happens after the response started", func() {
		s.buf.Reset()
		partial := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte("partial"))
			panic("late")
		})

		rec := s.do(partial, "p2")

		s.Equal(http.StatusAccepted, rec.Code)
		s.Equal("partial", rec.Body.String())
	})

	s.Run("should treat a body write as an implicit 200", func() {
		s.buf.Reset()
		partial := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("hi"))
			panic("late")
		})

		rec := s.do(partial, "p3")

		s.Equal(http.StatusOK, rec.Code)
		s.Equal("hi", rec.Body.String())
	})
}
