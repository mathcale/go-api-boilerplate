package handlers

import (
	"bytes"
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

type AuthHandlerTestSuite struct {
	suite.Suite
	response     Response
	loggerMock   *mocks.Logger
	signUpUCMock *mocks.SignUpUseCase
	signInUCMock *mocks.SignInUseCase

	handler Auth
}

func (s *AuthHandlerTestSuite) SetupTest() {
	s.loggerMock = new(mocks.Logger)
	s.response = NewResponse(s.loggerMock)
	s.signUpUCMock = new(mocks.SignUpUseCase)
	s.signInUCMock = new(mocks.SignInUseCase)

	s.handler = NewAuthHandler(s.response, s.signUpUCMock, s.signInUCMock)
}

func (s *AuthHandlerTestSuite) cleanMocks() {
	s.loggerMock.ExpectedCalls = nil
	s.loggerMock.Calls = nil
	s.signUpUCMock.ExpectedCalls = nil
	s.signUpUCMock.Calls = nil
	s.signInUCMock.ExpectedCalls = nil
	s.signInUCMock.Calls = nil
}

func TestAuthHandler(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func (s *AuthHandlerTestSuite) TestSignIn() {
	at := "any-jwt-access-token"
	rt := "any-jwt-refresh-token"

	s.Run("should return token for authenticated user", func() {
		defer s.cleanMocks()

		input := []byte(`{"email":"foo@example.com","password":"any-password"}`)

		r := httptest.NewRequest(http.MethodPost, "/v1/auth/signin", bytes.NewBuffer(input))
		w := httptest.NewRecorder()

		s.signInUCMock.
			On("Execute", r.Context(), mock.AnythingOfType("domain.User")).
			Return(&at, &rt, nil)

		s.handler.SignIn(w, r)

		res := w.Result()
		defer res.Body.Close()

		data, _ := io.ReadAll(res.Body)
		expected := `{"access_token":"any-jwt-access-token","refresh_token":"any-jwt-refresh-token"}`

		s.Equal(http.StatusOK, res.StatusCode)
		s.Equal(expected, strings.TrimSuffix(string(data), "\n"))
	})

	s.Run("should return parse error when loading input data", func() {
		defer s.cleanMocks()

		input := []byte(`im-not-a-valid-json-object`)

		r := httptest.NewRequest(http.MethodPost, "/v1/auth/signin", bytes.NewBuffer(input))
		w := httptest.NewRecorder()

		s.loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything)

		s.handler.SignIn(w, r)

		res := w.Result()

		s.Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("should return validation error when payload has missing value", func() {
		defer s.cleanMocks()

		input := []byte(`{"email":"foo@example.com"}`)

		r := httptest.NewRequest(http.MethodPost, "/v1/auth/signin", bytes.NewBuffer(input))
		w := httptest.NewRecorder()

		s.loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything)

		s.handler.SignIn(w, r)

		res := w.Result()

		s.Equal(http.StatusUnprocessableEntity, res.StatusCode)
	})

	s.Run("should return error when something goes wrong while authenticating", func() {
		defer s.cleanMocks()

		input := []byte(`{"email":"foo@example.com","password":"any-password"}`)

		r := httptest.NewRequest(http.MethodPost, "/v1/auth/signin", bytes.NewBuffer(input))
		w := httptest.NewRecorder()

		s.loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything)
		s.signInUCMock.
			On("Execute", r.Context(), mock.AnythingOfType("domain.User")).
			Return(nil, nil, errors.New("any-error"))

		s.handler.SignIn(w, r)

		res := w.Result()
		defer res.Body.Close()

		data, _ := io.ReadAll(res.Body)
		expected := `{"message":"any-error"}`

		s.Equal(http.StatusInternalServerError, res.StatusCode)
		s.Equal(expected, strings.TrimSuffix(string(data), "\n"))
	})
}

func (s *AuthHandlerTestSuite) TestSignUp() {
	s.Run("should create new user", func() {
		defer s.cleanMocks()

		input := []byte(`{"name":"any-name","email":"foo@example.com","password":"any-password"}`)

		r := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", bytes.NewBuffer(input))
		w := httptest.NewRecorder()

		s.signUpUCMock.
			On("Execute", r.Context(), mock.AnythingOfType("domain.User")).
			Return(nil)

		s.handler.SignUp(w, r)

		res := w.Result()

		s.Equal(http.StatusCreated, res.StatusCode)
	})

	s.Run("should return parse error when loading input data", func() {
		defer s.cleanMocks()

		input := []byte(`im-not-a-valid-json-object`)

		r := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", bytes.NewBuffer(input))
		w := httptest.NewRecorder()

		s.loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything)

		s.handler.SignUp(w, r)

		res := w.Result()

		s.Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("should return validation error when payload has missing value", func() {
		defer s.cleanMocks()

		input := []byte(`{"name":"any-name"}`)

		r := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", bytes.NewBuffer(input))
		w := httptest.NewRecorder()

		s.loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything)

		s.handler.SignUp(w, r)

		res := w.Result()

		s.Equal(http.StatusUnprocessableEntity, res.StatusCode)
	})

	s.Run("should return error when something goes wrong while creating user", func() {
		defer s.cleanMocks()

		input := []byte(`{"name":"any-name","email":"foo@example.com","password":"any-password"}`)

		r := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", bytes.NewBuffer(input))
		w := httptest.NewRecorder()

		s.loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything)
		s.signUpUCMock.
			On("Execute", r.Context(), mock.AnythingOfType("domain.User")).
			Return(errors.New("any-error"))

		s.handler.SignUp(w, r)

		res := w.Result()
		defer res.Body.Close()

		data, _ := io.ReadAll(res.Body)
		expected := `{"message":"any-error"}`

		s.Equal(http.StatusInternalServerError, res.StatusCode)
		s.Equal(expected, strings.TrimSuffix(string(data), "\n"))
	})
}
