package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers/dto"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

func newAuthHandler(useCases handlers.AuthUseCases) handlers.AuthHandler {
	response := handlers.NewResponse(mocks.NoopLogger{}, config.EnvironmentTest)
	return handlers.NewAuthHandler(response, useCases)
}

func TestAuthHandler_SignIn_Success(t *testing.T) {
	signIn := new(mocks.SignInUseCase)
	signIn.On("Execute", mock.Anything, "jane@example.com", "secret123").
		Return(&auth.TokenPair{AccessToken: "at", RefreshToken: "rt"}, nil)

	h := newAuthHandler(handlers.AuthUseCases{SignIn: signIn})

	body := `{"email":"jane@example.com","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/signin", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.SignIn(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var out dto.TokenPairOutput
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&out))
	assert.Equal(t, "at", out.AccessToken)
	assert.Equal(t, "rt", out.RefreshToken)
	signIn.AssertExpectations(t)
}

func TestAuthHandler_SignIn_ValidationError(t *testing.T) {
	signIn := new(mocks.SignInUseCase)

	h := newAuthHandler(handlers.AuthUseCases{SignIn: signIn})

	// Missing password fails struct validation before the use case is invoked.
	body := `{"email":"jane@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/signin", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.SignIn(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	signIn.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthHandler_SignIn_MalformedJSON(t *testing.T) {
	signIn := new(mocks.SignInUseCase)

	h := newAuthHandler(handlers.AuthUseCases{SignIn: signIn})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/signin", strings.NewReader("{"))
	rec := httptest.NewRecorder()

	h.SignIn(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	signIn.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything, mock.Anything)
}
