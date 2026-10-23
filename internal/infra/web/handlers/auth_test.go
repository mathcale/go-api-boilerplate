package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/infra/web/handlers/dto"
	"go-api-boilerplate/internal/infra/web/webcontext"
	"go-api-boilerplate/internal/pkg/apperror"
	"go-api-boilerplate/internal/tests/fixtures"
	"go-api-boilerplate/internal/tests/mocks"
)

type AuthHandlerTestSuite struct {
	suite.Suite

	handler AuthHandler

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

	response := NewResponse(mocks.NoopLogger{}, config.EnvironmentTest)
	s.handler = NewAuthHandler(response, AuthUseCases{
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

func (s *AuthHandlerTestSuite) SetupSubTest() {
	s.SetupTest()
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
	path := "/v1/auth/signin"

	s.Run("should return an access and refresh token pair on successful sign in", func() {
		defer s.cleanMocks()

		s.signInMock.On("Execute", mock.Anything, fixtures.WebEmail, fixtures.WebPassword).
			Return(fixtures.WebTokenPair(), nil)

		rec := s.serve(s.handler.SignIn, http.MethodPost, path, fixtures.WebSignInBody())

		s.Equal(http.StatusOK, rec.Code)

		var out dto.TokenPairOutput
		s.Require().NoError(json.NewDecoder(rec.Body).Decode(&out))
		s.Equal(fixtures.WebAccessToken, out.AccessToken)
		s.Equal(fixtures.WebRefreshToken, out.RefreshToken)
		s.signInMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the sign in payload fails validation", func() {
		defer s.cleanMocks()

		body := fixtures.WebSignInMissingPasswordBody()
		rec := s.serve(s.handler.SignIn, http.MethodPost, path, body)

		s.assertError(rec, http.StatusBadRequest, apperror.BE0002_INPUT_VALIDATION_FAILED)
		s.signInMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return bad request when the request body is malformed json", func() {
		defer s.cleanMocks()

		rec := s.serve(s.handler.SignIn, http.MethodPost, path, fixtures.WebMalformedBody)

		s.assertError(rec, http.StatusBadRequest, apperror.BE0001_INVALID_INPUT)
		s.signInMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the sign in use case fails", func() {
		defer s.cleanMocks()

		s.signInMock.On("Execute", mock.Anything, fixtures.WebEmail, fixtures.WebPassword).
			Return(nil, errors.New("invalid credentials"))

		rec := s.serve(s.handler.SignIn, http.MethodPost, path, fixtures.WebSignInBody())

		s.assertError(rec, http.StatusInternalServerError, "")
		s.signInMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestSignUp() {
	path := "/v1/auth/signup"

	s.Run("should create a new user and return its id and email on successful sign up", func() {
		defer s.cleanMocks()

		created := fixtures.WebCreatedUser()
		s.signUpMock.On("Execute", mock.Anything, mock.AnythingOfType("auth.SignUpInput")).
			Return(created, nil)

		rec := s.serve(s.handler.SignUp, http.MethodPost, path, fixtures.WebSignUpBody())

		s.Equal(http.StatusCreated, rec.Code)

		var out dto.SignUpOutput
		s.Require().NoError(json.NewDecoder(rec.Body).Decode(&out))
		s.Equal(created.ID.String(), out.ID)
		s.Equal(created.Email, out.Email)
		s.signUpMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the sign up payload fails validation", func() {
		defer s.cleanMocks()

		body := fixtures.WebSignUpMissingFieldsBody()
		rec := s.serve(s.handler.SignUp, http.MethodPost, path, body)

		s.assertError(rec, http.StatusBadRequest, apperror.BE0002_INPUT_VALIDATION_FAILED)
		s.signUpMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the sign up use case fails", func() {
		defer s.cleanMocks()

		s.signUpMock.On("Execute", mock.Anything, mock.AnythingOfType("auth.SignUpInput")).
			Return(nil, errors.New("user already exists"))

		rec := s.serve(s.handler.SignUp, http.MethodPost, path, fixtures.WebSignUpBody())

		s.assertError(rec, http.StatusInternalServerError, "")
		s.signUpMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestRefreshToken() {
	path := "/v1/auth/refresh-token"

	s.Run("should return a new token pair on successful refresh", func() {
		defer s.cleanMocks()

		s.refreshTokenMock.On("Execute", mock.Anything, fixtures.WebOldRefreshToken).
			Return(fixtures.WebTokenPair(), nil)

		body := fixtures.WebRefreshTokenBody(fixtures.WebOldRefreshToken)
		rec := s.serve(s.handler.RefreshToken, http.MethodPost, path, body)

		s.Equal(http.StatusOK, rec.Code)

		var out dto.TokenPairOutput
		s.Require().NoError(json.NewDecoder(rec.Body).Decode(&out))
		s.Equal(fixtures.WebAccessToken, out.AccessToken)
		s.refreshTokenMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the refresh token payload fails validation", func() {
		defer s.cleanMocks()

		rec := s.serve(s.handler.RefreshToken, http.MethodPost, path, fixtures.WebEmptyBody)

		s.assertError(rec, http.StatusBadRequest, apperror.BE0002_INPUT_VALIDATION_FAILED)
		s.refreshTokenMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the refresh token use case fails", func() {
		defer s.cleanMocks()

		s.refreshTokenMock.On("Execute", mock.Anything, fixtures.WebBadRefreshToken).
			Return(nil, errors.New("unauthorized"))

		body := fixtures.WebRefreshTokenBody(fixtures.WebBadRefreshToken)
		rec := s.serve(s.handler.RefreshToken, http.MethodPost, path, body)

		s.assertError(rec, http.StatusInternalServerError, "")
		s.refreshTokenMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestConfirmAccount() {
	path := "/v1/auth/confirm-account"

	s.Run("should return no content on successful account confirmation", func() {
		defer s.cleanMocks()

		userID := fixtures.WebUserID()
		s.confirmAccountMock.On("Execute", mock.Anything, userID, fixtures.WebCode).Return(nil)

		body := fixtures.WebUserCodeBody(userID.String(), fixtures.WebCode)
		rec := s.serve(s.handler.ConfirmAccount, http.MethodPost, path, body)

		s.Equal(http.StatusNoContent, rec.Code)
		s.confirmAccountMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the confirm account payload fails validation", func() {
		defer s.cleanMocks()

		rec := s.serve(s.handler.ConfirmAccount, http.MethodPost, path, fixtures.WebEmptyBody)

		s.assertError(rec, http.StatusBadRequest, apperror.BE0002_INPUT_VALIDATION_FAILED)
		s.confirmAccountMock.AssertNotCalled(
			s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything,
		)
	})

	s.Run("should return bad request when the user id is not a valid uuid", func() {
		defer s.cleanMocks()

		body := fixtures.WebUserCodeBody(fixtures.WebShortUserID, fixtures.WebCode)
		rec := s.serve(s.handler.ConfirmAccount, http.MethodPost, path, body)

		s.assertError(rec, http.StatusBadRequest, apperror.BE0002_INPUT_VALIDATION_FAILED)
		s.confirmAccountMock.AssertNotCalled(
			s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything,
		)
	})

	s.Run("should return internal server error when the confirm account use case fails", func() {
		defer s.cleanMocks()

		userID := fixtures.WebUserID()
		s.confirmAccountMock.On("Execute", mock.Anything, userID, fixtures.WebBadCode).
			Return(errors.New("invalid code"))

		body := fixtures.WebUserCodeBody(userID.String(), fixtures.WebBadCode)
		rec := s.serve(s.handler.ConfirmAccount, http.MethodPost, path, body)

		s.assertError(rec, http.StatusInternalServerError, "")
		s.confirmAccountMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestResendConfirmationCode() {
	path := "/v1/auth/resend-confirmation-code"

	s.Run("should return no content on successful confirmation code resend", func() {
		defer s.cleanMocks()

		s.resendMock.On("Execute", mock.Anything, fixtures.WebEmail).Return(nil)

		rec := s.serve(
			s.handler.ResendConfirmationCode, http.MethodPost, path, fixtures.WebEmailBody(),
		)

		s.Equal(http.StatusNoContent, rec.Code)
		s.resendMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the resend payload fails validation", func() {
		defer s.cleanMocks()

		rec := s.serve(
			s.handler.ResendConfirmationCode, http.MethodPost, path, fixtures.WebEmptyBody,
		)

		s.assertError(rec, http.StatusBadRequest, apperror.BE0002_INPUT_VALIDATION_FAILED)
		s.resendMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the resend use case fails", func() {
		defer s.cleanMocks()

		s.resendMock.On("Execute", mock.Anything, fixtures.WebEmail).Return(errors.New("boom"))

		rec := s.serve(
			s.handler.ResendConfirmationCode, http.MethodPost, path, fixtures.WebEmailBody(),
		)

		s.assertError(rec, http.StatusInternalServerError, "")
		s.resendMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestMe() {
	path := "/v1/auth/me"

	s.Run("should return the authenticated user's data", func() {
		defer s.cleanMocks()

		userID := fixtures.WebUserID()
		s.meMock.On("Execute", mock.Anything, userID).
			Return(fixtures.WebAuthenticatedUser(userID), nil)

		req := httptest.NewRequest(http.MethodGet, path, nil)
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

		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		s.handler.Me(rec, req)

		s.assertError(rec, http.StatusUnauthorized, apperror.BE0003_UNAUTHORIZED)
		s.meMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the me use case fails", func() {
		defer s.cleanMocks()

		userID := fixtures.WebUserID()
		s.meMock.On("Execute", mock.Anything, userID).Return(nil, errors.New("not found"))

		req := httptest.NewRequest(http.MethodGet, path, nil)
		req = req.WithContext(webcontext.WithUserID(req.Context(), userID))
		rec := httptest.NewRecorder()

		s.handler.Me(rec, req)

		s.assertError(rec, http.StatusInternalServerError, "")
		s.meMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestSetRecoveryCode() {
	path := "/v1/auth/recovery-code"

	s.Run("should return no content on successful recovery code creation", func() {
		defer s.cleanMocks()

		s.setRecoveryCodeMock.On("Execute", mock.Anything, fixtures.WebEmail).Return(nil)

		rec := s.serve(s.handler.SetRecoveryCode, http.MethodPost, path, fixtures.WebEmailBody())

		s.Equal(http.StatusNoContent, rec.Code)
		s.setRecoveryCodeMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the set recovery code payload fails validation", func() {
		defer s.cleanMocks()

		rec := s.serve(s.handler.SetRecoveryCode, http.MethodPost, path, fixtures.WebEmptyBody)

		s.assertError(rec, http.StatusBadRequest, apperror.BE0002_INPUT_VALIDATION_FAILED)
		s.setRecoveryCodeMock.AssertNotCalled(s.T(), "Execute", mock.Anything, mock.Anything)
	})

	s.Run("should return internal server error when the set recovery code use case fails", func() {
		defer s.cleanMocks()

		s.setRecoveryCodeMock.On("Execute", mock.Anything, fixtures.WebEmail).
			Return(errors.New("boom"))

		rec := s.serve(s.handler.SetRecoveryCode, http.MethodPost, path, fixtures.WebEmailBody())

		s.assertError(rec, http.StatusInternalServerError, "")
		s.setRecoveryCodeMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestValidateRecoveryCode() {
	path := "/v1/auth/recovery-code/validate"

	s.Run("should return no content on successful recovery code validation", func() {
		defer s.cleanMocks()

		userID := fixtures.WebUserID()
		s.validateRecoveryCodeMock.On("Execute", mock.Anything, userID, fixtures.WebCode).
			Return(nil)

		body := fixtures.WebUserCodeBody(userID.String(), fixtures.WebCode)
		rec := s.serve(s.handler.ValidateRecoveryCode, http.MethodPost, path, body)

		s.Equal(http.StatusNoContent, rec.Code)
		s.validateRecoveryCodeMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the validate recovery code payload is invalid", func() {
		defer s.cleanMocks()

		rec := s.serve(s.handler.ValidateRecoveryCode, http.MethodPost, path, fixtures.WebEmptyBody)

		s.assertError(rec, http.StatusBadRequest, apperror.BE0002_INPUT_VALIDATION_FAILED)
		s.validateRecoveryCodeMock.AssertNotCalled(
			s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything,
		)
	})

	s.Run("should return bad request when the user id is not a valid uuid", func() {
		defer s.cleanMocks()

		body := fixtures.WebUserCodeBody(fixtures.WebInvalidUserID, fixtures.WebCode)
		rec := s.serve(s.handler.ValidateRecoveryCode, http.MethodPost, path, body)

		s.assertError(rec, http.StatusBadRequest, apperror.BE0002_INPUT_VALIDATION_FAILED)
		s.validateRecoveryCodeMock.AssertNotCalled(
			s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything,
		)
	})

	s.Run("should return internal server error when the validate recovery code fails", func() {
		defer s.cleanMocks()

		userID := fixtures.WebUserID()
		s.validateRecoveryCodeMock.On("Execute", mock.Anything, userID, fixtures.WebBadCode).
			Return(errors.New("boom"))

		body := fixtures.WebUserCodeBody(userID.String(), fixtures.WebBadCode)
		rec := s.serve(s.handler.ValidateRecoveryCode, http.MethodPost, path, body)

		s.assertError(rec, http.StatusInternalServerError, "")
		s.validateRecoveryCodeMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) TestUpdatePassword() {
	path := "/v1/auth/password"

	s.Run("should return no content on successful password update", func() {
		defer s.cleanMocks()

		userID := fixtures.WebUserID()
		s.updatePasswordMock.On(
			"Execute", mock.Anything, userID, fixtures.WebCode, fixtures.WebNewPassword,
		).Return(nil)

		body := fixtures.WebUpdatePasswordBody(userID.String(), fixtures.WebCode)
		rec := s.serve(s.handler.UpdatePassword, http.MethodPut, path, body)

		s.Equal(http.StatusNoContent, rec.Code)
		s.updatePasswordMock.AssertExpectations(s.T())
	})

	s.Run("should return bad request when the update password payload fails validation", func() {
		defer s.cleanMocks()

		rec := s.serve(s.handler.UpdatePassword, http.MethodPut, path, fixtures.WebEmptyBody)

		s.assertError(rec, http.StatusBadRequest, apperror.BE0002_INPUT_VALIDATION_FAILED)
		s.updatePasswordMock.AssertNotCalled(
			s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		)
	})

	s.Run("should return bad request when the user id is not a valid uuid", func() {
		defer s.cleanMocks()

		body := fixtures.WebUpdatePasswordBody(fixtures.WebInvalidUserID, fixtures.WebCode)
		rec := s.serve(s.handler.UpdatePassword, http.MethodPut, path, body)

		s.assertError(rec, http.StatusBadRequest, apperror.BE0002_INPUT_VALIDATION_FAILED)
		s.updatePasswordMock.AssertNotCalled(
			s.T(), "Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		)
	})

	s.Run("should return internal server error when the update password use case fails", func() {
		defer s.cleanMocks()

		userID := fixtures.WebUserID()
		s.updatePasswordMock.On(
			"Execute", mock.Anything, userID, fixtures.WebCode, fixtures.WebNewPassword,
		).Return(errors.New("boom"))

		body := fixtures.WebUpdatePasswordBody(userID.String(), fixtures.WebCode)
		rec := s.serve(s.handler.UpdatePassword, http.MethodPut, path, body)

		s.assertError(rec, http.StatusInternalServerError, "")
		s.updatePasswordMock.AssertExpectations(s.T())
	})
}

func (s *AuthHandlerTestSuite) serve(
	fn http.HandlerFunc,
	method, path, body string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()

	fn(rec, req)

	return rec
}

func (s *AuthHandlerTestSuite) assertError(
	rec *httptest.ResponseRecorder,
	status int,
	code apperror.BusinessCode,
) {
	s.T().Helper()

	s.Equal(status, rec.Code)

	var out dto.ErrorOutput
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&out))

	if code == "" {
		s.Nil(out.Code)
		return
	}

	s.Require().NotNil(out.Code)
	s.Equal(code, *out.Code)
}
