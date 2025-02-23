package gateways

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
	authuc "github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

type SignUpGatewayTestSuite struct {
	suite.Suite
	ctx          context.Context
	userRepoMock *mocks.UserRepositoryMock
	gw           authuc.SignUpGateway
}

func (s *SignUpGatewayTestSuite) SetupTest() {
	s.userRepoMock = new(mocks.UserRepositoryMock)
	s.gw = NewSignUpGateway(s.userRepoMock)
}

func (s *SignUpGatewayTestSuite) cleanMocks() {
	s.userRepoMock.ExpectedCalls = nil
	s.userRepoMock.Calls = nil
}

func TestSignUpGateway(t *testing.T) {
	suite.Run(t, new(SignUpGatewayTestSuite))
}

func (s *SignUpGatewayTestSuite) TestUserExists() {
	fixt := fixtures.NewUserModelComplete()
	truePtr := true

	s.Run("should find user and return it", func() {
		defer s.cleanMocks()

		s.userRepoMock.On("Exists", s.ctx, fixt.Email).Return(&truePtr, nil)

		exists, err := s.gw.UserExists(s.ctx, fixt.Email)

		s.NoError(err)
		s.NotNil(exists)
		s.True(*exists)
	})

	s.Run("should return error while retrieving user", func() {
		defer s.cleanMocks()

		s.userRepoMock.On("Exists", s.ctx, fixt.Email).Return(nil, errors.New("any-repo-error"))

		exists, err := s.gw.UserExists(s.ctx, fixt.Email)

		s.Error(err)
		s.ErrorContains(err, "any-repo-error")
		s.Nil(exists)
	})
}

func (s *SignUpGatewayTestSuite) TestSaveUser() {
	fixt := fixtures.NewUserDomain()

	s.Run("should save user", func() {
		defer s.cleanMocks()

		s.userRepoMock.On("Save", s.ctx, fixt.ToPartialModel()).Return(nil)

		err := s.gw.SaveUser(s.ctx, fixt)

		s.NoError(err)
	})

	s.Run("should return error while saving user", func() {
		defer s.cleanMocks()

		s.userRepoMock.On("Save", s.ctx, fixt.ToPartialModel()).Return(errors.New("any-repo-error"))

		err := s.gw.SaveUser(s.ctx, fixt)

		s.Error(err)
		s.ErrorContains(err, "any-repo-error")
	})
}
