package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/tests"
	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

type SignUpUseCaseTestSuite struct {
	suite.Suite
	ctx         context.Context
	loggerMock  *mocks.Logger
	passwdMock  *mocks.BcryptPassword
	gatewayMock *mocks.SignUpGateway

	uc SignUpUseCase
}

func (s *SignUpUseCaseTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.loggerMock = new(mocks.Logger)
	s.passwdMock = new(mocks.BcryptPassword)
	s.gatewayMock = new(mocks.SignUpGateway)

	s.loggerMock.On("Debug", mock.Anything, mock.Anything)

	s.uc = NewSignUpUseCase(s.loggerMock, s.passwdMock, s.gatewayMock)
}

func (s *SignUpUseCaseTestSuite) cleanMocks() {
	s.loggerMock.ExpectedCalls = nil
	s.loggerMock.Calls = nil
	s.passwdMock.ExpectedCalls = nil
	s.passwdMock.Calls = nil
	s.gatewayMock.ExpectedCalls = nil
	s.gatewayMock.Calls = nil
}

func TestSignUpUseCase(t *testing.T) {
	suite.Run(t, new(SignUpUseCaseTestSuite))
}

func (s *SignUpUseCaseTestSuite) TestExecute() {
	user := fixtures.NewUserDomainComplete()
	input := fixtures.NewUserDomain()

	s.Run("should create new user", func() {
		defer s.cleanMocks()

		s.gatewayMock.On("UserExists", s.ctx, input.Email).Return(tests.ToBoolPointer(false), nil)
		s.passwdMock.On("Hash", input.Password).Return(&user.Password, nil)
		s.gatewayMock.On("SaveUser", s.ctx, mock.AnythingOfType("domain.User")).Return(nil)

		err := s.uc.Execute(s.ctx, input)

		s.NoError(err)
	})

	s.Run("should return error while checking if user exists", func() {
		defer s.cleanMocks()

		s.loggerMock.On("Debug", mock.Anything, mock.Anything)
		s.gatewayMock.On("UserExists", s.ctx, input.Email).Return(tests.ToBoolPointer(false), errors.New("any-error"))

		err := s.uc.Execute(s.ctx, input)

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})

	s.Run("should return error when user already exists", func() {
		defer s.cleanMocks()

		s.loggerMock.On("Debug", mock.Anything, mock.Anything)
		s.gatewayMock.On("UserExists", s.ctx, input.Email).Return(tests.ToBoolPointer(true), nil)

		err := s.uc.Execute(s.ctx, input)

		s.Error(err)
	})

	s.Run("should return error while hashing password", func() {
		defer s.cleanMocks()

		s.loggerMock.On("Debug", mock.Anything, mock.Anything)
		s.gatewayMock.On("UserExists", s.ctx, input.Email).Return(tests.ToBoolPointer(false), nil)
		s.passwdMock.On("Hash", input.Password).Return(nil, errors.New("any-error"))

		err := s.uc.Execute(s.ctx, input)

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})

	s.Run("should return error while saving user", func() {
		defer s.cleanMocks()

		s.loggerMock.On("Debug", mock.Anything, mock.Anything)
		s.gatewayMock.On("UserExists", s.ctx, input.Email).Return(tests.ToBoolPointer(false), nil)
		s.passwdMock.On("Hash", input.Password).Return(&user.Password, nil)
		s.gatewayMock.On("SaveUser", s.ctx, mock.AnythingOfType("domain.User")).Return(errors.New("any-error"))

		err := s.uc.Execute(s.ctx, input)

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})
}
