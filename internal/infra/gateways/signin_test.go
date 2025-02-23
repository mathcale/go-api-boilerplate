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

type SignInGatewayTestSuite struct {
	suite.Suite
	ctx          context.Context
	userRepoMock *mocks.UserRepositoryMock
	gw           authuc.SignInGateway
}

func (s *SignInGatewayTestSuite) SetupTest() {
	s.userRepoMock = new(mocks.UserRepositoryMock)
	s.gw = NewSignInGateway(s.userRepoMock)
}

func (s *SignInGatewayTestSuite) cleanMocks() {
	s.userRepoMock.ExpectedCalls = nil
	s.userRepoMock.Calls = nil
}

func TestSignInGateway(t *testing.T) {
	suite.Run(t, new(SignInGatewayTestSuite))
}

func (s *SignInGatewayTestSuite) TestGetUser() {
	fixt := fixtures.NewUserModelComplete()

	s.Run("should find user and return it", func() {
		defer s.cleanMocks()

		s.userRepoMock.On("Get", s.ctx, fixt.Email).Return(&fixt, nil)

		user, err := s.gw.GetUser(s.ctx, fixt.Email)

		s.NoError(err)
		s.NotNil(user)
		s.Equal(fixt.Email, user.Email)
	})

	s.Run("should return error while retrieving user", func() {
		defer s.cleanMocks()

		s.userRepoMock.On("Get", s.ctx, fixt.Email).Return(nil, errors.New("any-repo-error"))

		user, err := s.gw.GetUser(s.ctx, fixt.Email)

		s.Error(err)
		s.ErrorContains(err, "any-repo-error")
		s.Nil(user)
	})
}
