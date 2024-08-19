package counter

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
	"github.com/mathcale/go-api-boilerplate/internal/web/handlers"
)

type CounterHandlerTestSuite struct {
	suite.Suite
	ResponseHandler    handlers.ResponseHandler
	CounterUseCaseMock *mocks.CounterUseCaseMock

	CounterHandler CounterHandler
}

func (s *CounterHandlerTestSuite) SetupTest() {
	s.ResponseHandler = handlers.NewResponseHandler()
	s.CounterUseCaseMock = new(mocks.CounterUseCaseMock)

	s.CounterHandler = NewCounterHandler(s.ResponseHandler, s.CounterUseCaseMock)
}

func (s *CounterHandlerTestSuite) cleanMocks() {
	s.CounterUseCaseMock.ExpectedCalls = nil
	s.CounterUseCaseMock.Calls = nil
}

func TestCounterHandler(t *testing.T) {
	suite.Run(t, new(CounterHandlerTestSuite))
}

func (s *CounterHandlerTestSuite) TestHandle() {
	s.Run("should return updated counter value", func() {
		defer s.cleanMocks()

		r := httptest.NewRequest(http.MethodGet, "/counter", nil)
		w := httptest.NewRecorder()

		s.CounterUseCaseMock.On("Execute").Return(1, nil)

		s.CounterHandler.Handle(w, r)

		res := w.Result()
		defer res.Body.Close()

		data, _ := io.ReadAll(res.Body)
		expected := `{"counter":1}`

		s.Equal(http.StatusOK, res.StatusCode)
		s.Equal(expected, strings.TrimSuffix(string(data), "\n"))
	})

	s.Run("should handle error from use case when something goes wrong", func() {
		defer s.cleanMocks()

		r := httptest.NewRequest(http.MethodGet, "/counter", nil)
		w := httptest.NewRecorder()

		s.CounterUseCaseMock.On("Execute").Return(0, errors.New("any-error"))

		s.CounterHandler.Handle(w, r)

		res := w.Result()
		defer res.Body.Close()

		data, _ := io.ReadAll(res.Body)
		expected := `{"message":"any-error"}`

		s.Equal(http.StatusInternalServerError, res.StatusCode)
		s.Equal(expected, strings.TrimSuffix(string(data), "\n"))
	})
}
