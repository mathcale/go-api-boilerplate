package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

type CounterHandlerTestSuite struct {
	suite.Suite
	loggerMock    *mocks.Logger
	counterUCMock *mocks.CounterUseCase
	response      Response

	handler Counter
}

func (s *CounterHandlerTestSuite) SetupTest() {
	s.loggerMock = new(mocks.Logger)
	s.counterUCMock = new(mocks.CounterUseCase)
	s.response = NewResponse(s.loggerMock)

	s.handler = NewCounterHandler(s.response, s.counterUCMock)
}

func (s *CounterHandlerTestSuite) cleanMocks() {
	s.counterUCMock.ExpectedCalls = nil
	s.counterUCMock.Calls = nil
}

func TestCounterHandler(t *testing.T) {
	suite.Run(t, new(CounterHandlerTestSuite))
}

func (s *CounterHandlerTestSuite) TestHandle() {
	s.Run("should return updated counter value", func() {
		defer s.cleanMocks()

		r := httptest.NewRequest(http.MethodGet, "/counter", nil)
		w := httptest.NewRecorder()

		s.counterUCMock.On("Execute").Return(1, nil)

		s.handler.Count(w, r)

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

		s.counterUCMock.On("Execute").Return(0, errors.New("any-error"))
		s.loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything)

		s.handler.Count(w, r)

		res := w.Result()
		defer res.Body.Close()

		data, _ := io.ReadAll(res.Body)
		expected := `{"code":null,"message":"any-error"}`

		s.Equal(http.StatusInternalServerError, res.StatusCode)
		s.Equal(expected, strings.TrimSuffix(string(data), "\n"))
	})
}
