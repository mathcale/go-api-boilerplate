package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

type SignInUseCaseTestSuite struct {
	suite.Suite
	ctx         context.Context
	loggerMock  *mocks.Logger
	passwdMock  *mocks.BcryptPassword
	jwtMock     *mocks.JWTAuth
	gatewayMock *mocks.SignInGateway

	uc SignInUseCase
}

func (s *SignInUseCaseTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.loggerMock = new(mocks.Logger)
	s.passwdMock = new(mocks.BcryptPassword)
	s.jwtMock = new(mocks.JWTAuth)
	s.gatewayMock = new(mocks.SignInGateway)

	s.loggerMock.On("Debug", mock.Anything, mock.Anything)

	s.uc = NewSignInUseCase(s.loggerMock, s.passwdMock, s.jwtMock, s.gatewayMock)
}

func (s *SignInUseCaseTestSuite) cleanMocks() {
	s.loggerMock.ExpectedCalls = nil
	s.loggerMock.Calls = nil
	s.passwdMock.ExpectedCalls = nil
	s.passwdMock.Calls = nil
	s.jwtMock.ExpectedCalls = nil
	s.jwtMock.Calls = nil
	s.gatewayMock.ExpectedCalls = nil
	s.gatewayMock.Calls = nil
}

func TestSignInUseCase(t *testing.T) {
	suite.Run(t, new(SignInUseCaseTestSuite))
}

func (s *SignInUseCaseTestSuite) TestExecute() {
	user := fixtures.NewUserDomainComplete()
	input := fixtures.NewUserDomain()
	token := fixtures.NewToken()

	s.Run("should sign in and return valid token", func() {
		defer s.cleanMocks()

		s.gatewayMock.On("GetUser", s.ctx, input.Email).Return(&user, nil)
		s.passwdMock.On("Verify", input.Password, user.Password).Return(nil)
		s.jwtMock.On("IssueAccessToken", user.ID.String()).Return(&token, nil)
		s.jwtMock.On("IssueRefreshToken", user.ID.String()).Return(&token, nil)

		at, rt, err := s.uc.Execute(s.ctx, input)

		s.NoError(err)
		s.Equal(token.Token, at)
		s.Equal(token.Token, rt)
	})

	s.Run("should return error when finding user", func() {
		defer s.cleanMocks()

		s.gatewayMock.On("GetUser", s.ctx, input.Email).Return(nil, errors.New("any-error"))

		at, rt, err := s.uc.Execute(s.ctx, input)

		s.Error(err)
		s.ErrorContains(err, "any-error")
		s.Nil(at)
		s.Nil(rt)
	})

	s.Run("should return error while validating password", func() {
		defer s.cleanMocks()

		s.gatewayMock.On("GetUser", s.ctx, input.Email).Return(&user, nil)
		s.passwdMock.On("Verify", input.Password, user.Password).Return(errors.New("any-error"))

		at, rt, err := s.uc.Execute(s.ctx, input)

		s.Error(err)
		s.ErrorContains(err, "any-error")
		s.Nil(at)
		s.Nil(rt)
	})

	s.Run("should return error while issuing access token", func() {
		defer s.cleanMocks()

		s.gatewayMock.On("GetUser", s.ctx, input.Email).Return(&user, nil)
		s.passwdMock.On("Verify", input.Password, user.Password).Return(nil)
		s.jwtMock.On("IssueAccessToken", user.ID.String()).Return(nil, errors.New("any-error"))

		at, rt, err := s.uc.Execute(s.ctx, input)

		s.Error(err)
		s.ErrorContains(err, "any-error")
		s.Nil(at)
		s.Nil(rt)
	})

	s.Run("should return error while issuing refresh token", func() {
		defer s.cleanMocks()

		s.gatewayMock.On("GetUser", s.ctx, input.Email).Return(&user, nil)
		s.passwdMock.On("Verify", input.Password, user.Password).Return(nil)
		s.jwtMock.On("IssueAccessToken", user.ID.String()).Return(&token, nil)
		s.jwtMock.On("IssueRefreshToken", user.ID.String()).Return(nil, errors.New("any-error"))

		at, rt, err := s.uc.Execute(s.ctx, input)

		s.Error(err)
		s.ErrorContains(err, "any-error")
		s.Nil(at)
		s.Nil(rt)
	})
}
