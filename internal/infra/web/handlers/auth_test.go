package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers/dto"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/webcontext"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

type AuthHandlerTestSuite struct {
	suite.Suite

	handler handlers.AuthHandler

	signInMock               *mocks.SignInUseCase
	signUpMock               *mocks.SignUpUseCase
	refreshTokenMock         *mocks.RefreshTokenUseCase
	confirmAccountMock       *mocks.ConfirmAccountUseCase
	resendMock               *mocks.ResendConfirmationCodeUseCase
	meMock                   *mocks.MeUseCase
	setRecoveryCodeMock      *mocks.SetRecoveryCodeUseCase
	validateRecoveryCodeMock *mocks.ValidateRecoveryCodeUseCase
	updatePasswordMock       *mocks.UpdatePasswordUseCase
}

func (s *AuthHandlerTestSuite) SetupTest() {
	s.signInMock = new(mocks.SignInUseCase)
	s.signUpMock = new(mocks.SignUpUseCase)
	s.refreshTokenMock = new(mocks.RefreshTokenUseCase)
	s.confirmAccountMock = new(mocks.ConfirmAccountUseCase)
	s.resendMock = new(mocks.ResendConfirmationCodeUseCase)
	s.meMock = new(mocks.MeUseCase)
	s.setRecoveryCodeMock = new(mocks.SetRecoveryCodeUseCase)
	s.validateRecoveryCodeMock = new(mocks.ValidateRecoveryCodeUseCase)
	s.updatePasswordMock = new(mocks.UpdatePasswordUseCase)

	response := handlers.NewResponse(mocks.NoopLogger{}, config.EnvironmentTest)
	s.handler = handlers.NewAuthHandler(response, handlers.AuthUseCases{
		SignIn:               s.signInMock,
		SignUp:               s.signUpMock,
		RefreshToken:         s.refreshTokenMock,
		ConfirmAccount:       s.confirmAccountMock,
		ResendConfirmation:   s.resendMock,
		Me:                   s.meMock,
		SetRecoveryCode:      s.setRecoveryCodeMock,
		ValidateRecoveryCode: s.validateRecoveryCodeMock,
		UpdatePassword:       s.updatePasswordMock,
	})
}

func (s *AuthHandlerTestSuite) cleanMocks() {
	s.signInMock.ExpectedCalls = nil
	s.signInMock.Calls = nil
	s.signUpMock.ExpectedCalls = nil
	s.signUpMock.Calls = nil
	s.refreshTokenMock.ExpectedCalls = nil
	s.refreshTokenMock.Calls = nil
	s.confirmAccountMock.ExpectedCalls = nil
	s.confirmAccountMock.Calls = nil
	s.resendMock.ExpectedCalls = nil
	s.resendMock.Calls = nil
	s.meMock.ExpectedCalls = nil
	s.meMock.Calls = nil
	s.setRecoveryCodeMock.ExpectedCalls = nil
	s.setRecoveryCodeMock.Calls = nil
	s.validateRecoveryCodeMock.ExpectedCalls = nil
	s.validateRecoveryCodeMock.Calls = nil
	s.updatePasswordMock.ExpectedCalls = nil
	s.updatePasswordMock.Calls = nil
}

func TestAuthHandler(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func (s *AuthHandlerTestSuite) TestSignIn() {
	s.Run("should return an access and refresh token pair on successful sign in", func() {
		defer s.cleanMocks()

		s.signInMock.On("Execute", mock.Anything, "jane@example.com", "secret123").
			Return(&auth.TokenPair{AccessToken: "at", RefreshToken: "rt"}, nil)

		body := `{"email":"jane@example.com","password":"secret123"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/signin", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.SignIn(rec, req)

		s.Equal(http.StatusOK, rec.Code)

		var out dto.TokenPairOutput
		s.Require().NoError(json.NewDecoder(rec.Body).Decode(&out))
		s.Equal("at", out.AccessToken)
		s.Equal("rt", out.RefreshToken)
		s.signInMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the sign in payload fails validation", func() {
		defer s.cleanMocks()

		body := `{"email":"jane@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/signin", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.SignIn(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
		s.signInMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return bad request when the request body is malformed json", func() {
		defer s.cleanMocks()

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/signin", strings.NewReader("{"))
		rec := httptest.NewRecorder()

		s.handler.SignIn(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
		s.signInMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the sign in use case fails", func() {
		defer s.cleanMocks()

		s.signInMock.On("Execute", mock.Anything, "jane@example.com", "secret123").
			Return(nil, errors.New("invalid credentials"))

		body := `{"email":"jane@example.com","password":"secret123"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/signin", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.SignIn(rec, req)

		s.Equal(http.StatusInternalServerError, rec.Code)
		s.signInMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestSignUp() {
	s.Run("should create a new user and return its id and email on successful sign up", func() {
		defer s.cleanMocks()

		created := &user.User{ID: uuid.New(), Name: "Jane", Surname: "Doe", Email: "jane@example.com", Active: false}
		s.signUpMock.On("Execute", mock.Anything, mock.AnythingOfType("auth.SignUpInput")).Return(created, nil)

		body := `{"name":"Jane","surname":"Doe","email":"jane@example.com","password":"secret123"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.SignUp(rec, req)

		s.Equal(http.StatusCreated, rec.Code)

		var out dto.SignUpOutput
		s.Require().NoError(json.NewDecoder(rec.Body).Decode(&out))
		s.Equal(created.ID.String(), out.ID)
		s.Equal(created.Email, out.Email)
		s.signUpMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the sign up payload fails validation", func() {
		defer s.cleanMocks()

		body := `{"name":"Jane"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.SignUp(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
		s.signUpMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the sign up use case fails", func() {
		defer s.cleanMocks()

		s.signUpMock.On("Execute", mock.Anything, mock.AnythingOfType("auth.SignUpInput")).
			Return(nil, errors.New("user already exists"))

		body := `{"name":"Jane","surname":"Doe","email":"jane@example.com","password":"secret123"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.SignUp(rec, req)

		s.Equal(http.StatusInternalServerError, rec.Code)
		s.signUpMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestRefreshToken() {
	s.Run("should return a new token pair on successful refresh", func() {
		defer s.cleanMocks()

		s.refreshTokenMock.On("Execute", mock.Anything, "old-refresh-token").
			Return(&auth.TokenPair{AccessToken: "at", RefreshToken: "rt"}, nil)

		body := `{"refresh_token":"old-refresh-token"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh-token", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.RefreshToken(rec, req)

		s.Equal(http.StatusOK, rec.Code)

		var out dto.TokenPairOutput
		s.Require().NoError(json.NewDecoder(rec.Body).Decode(&out))
		s.Equal("at", out.AccessToken)
		s.refreshTokenMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the refresh token payload fails validation", func() {
		defer s.cleanMocks()

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh-token", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()

		s.handler.RefreshToken(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
		s.refreshTokenMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the refresh token use case fails", func() {
		defer s.cleanMocks()

		s.refreshTokenMock.On("Execute", mock.Anything, "bad-token").Return(nil, errors.New("unauthorized"))

		body := `{"refresh_token":"bad-token"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh-token", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.RefreshToken(rec, req)

		s.Equal(http.StatusInternalServerError, rec.Code)
		s.refreshTokenMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestConfirmAccount() {
	s.Run("should return no content on successful account confirmation", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.confirmAccountMock.On("Execute", mock.Anything, userID, "the-code").Return(nil)

		body := `{"user_id":"` + userID.String() + `","code":"the-code"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/confirm-account", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.ConfirmAccount(rec, req)

		s.Equal(http.StatusNoContent, rec.Code)
		s.confirmAccountMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the confirm account payload fails validation", func() {
		defer s.cleanMocks()

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/confirm-account", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()

		s.handler.ConfirmAccount(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
		s.confirmAccountMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return bad request when the user id is not a valid uuid", func() {
		defer s.cleanMocks()

		body := `{"user_id":"11111111-1111-1111-1111-11111111111","code":"the-code"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/confirm-account", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.ConfirmAccount(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
		s.confirmAccountMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the confirm account use case fails", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.confirmAccountMock.On("Execute", mock.Anything, userID, "bad-code").Return(errors.New("invalid code"))

		body := `{"user_id":"` + userID.String() + `","code":"bad-code"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/confirm-account", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.ConfirmAccount(rec, req)

		s.Equal(http.StatusInternalServerError, rec.Code)
		s.confirmAccountMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestResendConfirmationCode() {
	s.Run("should return no content on successful confirmation code resend", func() {
		defer s.cleanMocks()

		s.resendMock.On("Execute", mock.Anything, "jane@example.com").Return(nil)

		body := `{"email":"jane@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/resend-confirmation-code", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.ResendConfirmationCode(rec, req)

		s.Equal(http.StatusNoContent, rec.Code)
		s.resendMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the resend confirmation code payload fails validation", func() {
		defer s.cleanMocks()

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/resend-confirmation-code", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()

		s.handler.ResendConfirmationCode(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
		s.resendMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the resend confirmation code use case fails", func() {
		defer s.cleanMocks()

		s.resendMock.On("Execute", mock.Anything, "jane@example.com").Return(errors.New("boom"))

		body := `{"email":"jane@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/resend-confirmation-code", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.ResendConfirmationCode(rec, req)

		s.Equal(http.StatusInternalServerError, rec.Code)
		s.resendMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestMe() {
	s.Run("should return the authenticated user's data", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		found := &user.User{ID: userID, Name: "Jane", Email: "jane@example.com"}
		s.meMock.On("Execute", mock.Anything, userID).Return(found, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
		req = req.WithContext(webcontext.WithUserID(req.Context(), userID))
		rec := httptest.NewRecorder()

		s.handler.Me(rec, req)

		s.Equal(http.StatusOK, rec.Code)

		var out dto.MeOutput
		s.Require().NoError(json.NewDecoder(rec.Body).Decode(&out))
		s.Equal(userID.String(), out.ID)
		s.meMock.AssertExpectations(s.T())
	})

	s.Run("should return unauthorized when there is no user id in the request context", func() {
		defer s.cleanMocks()

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
		rec := httptest.NewRecorder()

		s.handler.Me(rec, req)

		s.Equal(http.StatusUnauthorized, rec.Code)
		s.meMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the me use case fails", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.meMock.On("Execute", mock.Anything, userID).Return(nil, errors.New("not found"))

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
		req = req.WithContext(webcontext.WithUserID(req.Context(), userID))
		rec := httptest.NewRecorder()

		s.handler.Me(rec, req)

		s.Equal(http.StatusInternalServerError, rec.Code)
		s.meMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestSetRecoveryCode() {
	s.Run("should return no content on successful recovery code creation", func() {
		defer s.cleanMocks()

		s.setRecoveryCodeMock.On("Execute", mock.Anything, "jane@example.com").Return(nil)

		body := `{"email":"jane@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/recovery-code", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.SetRecoveryCode(rec, req)

		s.Equal(http.StatusNoContent, rec.Code)
		s.setRecoveryCodeMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the set recovery code payload fails validation", func() {
		defer s.cleanMocks()

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/recovery-code", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()

		s.handler.SetRecoveryCode(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
		s.setRecoveryCodeMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the set recovery code use case fails", func() {
		defer s.cleanMocks()

		s.setRecoveryCodeMock.On("Execute", mock.Anything, "jane@example.com").Return(errors.New("boom"))

		body := `{"email":"jane@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/recovery-code", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.SetRecoveryCode(rec, req)

		s.Equal(http.StatusInternalServerError, rec.Code)
		s.setRecoveryCodeMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestValidateRecoveryCode() {
	s.Run("should return no content on successful recovery code validation", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.validateRecoveryCodeMock.On("Execute", mock.Anything, userID, "the-code").Return(nil)

		body := `{"user_id":"` + userID.String() + `","code":"the-code"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/recovery-code/validate", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.ValidateRecoveryCode(rec, req)

		s.Equal(http.StatusNoContent, rec.Code)
		s.validateRecoveryCodeMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the validate recovery code payload fails validation", func() {
		defer s.cleanMocks()

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/recovery-code/validate", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()

		s.handler.ValidateRecoveryCode(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
		s.validateRecoveryCodeMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return bad request when the user id is not a valid uuid", func() {
		defer s.cleanMocks()

		body := `{"user_id":"not-a-uuid","code":"the-code"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/recovery-code/validate", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.ValidateRecoveryCode(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
		s.validateRecoveryCodeMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the validate recovery code use case fails", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.validateRecoveryCodeMock.On("Execute", mock.Anything, userID, "bad-code").Return(errors.New("boom"))

		body := `{"user_id":"` + userID.String() + `","code":"bad-code"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/recovery-code/validate", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.ValidateRecoveryCode(rec, req)

		s.Equal(http.StatusInternalServerError, rec.Code)
		s.validateRecoveryCodeMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestUpdatePassword() {
	s.Run("should return no content on successful password update", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.updatePasswordMock.On("Execute", mock.Anything, userID, "the-code", "new-password123").Return(nil)

		body := `{"user_id":"` + userID.String() + `","code":"the-code","new_password":"new-password123"}`
		req := httptest.NewRequest(http.MethodPut, "/v1/auth/password", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.UpdatePassword(rec, req)

		s.Equal(http.StatusNoContent, rec.Code)
		s.updatePasswordMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the update password payload fails validation", func() {
		defer s.cleanMocks()

		req := httptest.NewRequest(http.MethodPut, "/v1/auth/password", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()

		s.handler.UpdatePassword(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
		s.updatePasswordMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return bad request when the user id is not a valid uuid", func() {
		defer s.cleanMocks()

		body := `{"user_id":"not-a-uuid","code":"the-code","new_password":"new-password123"}`
		req := httptest.NewRequest(http.MethodPut, "/v1/auth/password", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.UpdatePassword(rec, req)

		s.Equal(http.StatusBadRequest, rec.Code)
		s.updatePasswordMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the update password use case fails", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.updatePasswordMock.On("Execute", mock.Anything, userID, "the-code", "new-password123").
			Return(errors.New("boom"))

		body := `{"user_id":"` + userID.String() + `","code":"the-code","new_password":"new-password123"}`
		req := httptest.NewRequest(http.MethodPut, "/v1/auth/password", strings.NewReader(body))
		rec := httptest.NewRecorder()

		s.handler.UpdatePassword(rec, req)

		s.Equal(http.StatusInternalServerError, rec.Code)
		s.updatePasswordMock.AssertExpectations(s.T())
	})
}
