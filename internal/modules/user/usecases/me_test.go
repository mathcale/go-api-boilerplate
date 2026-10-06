package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/core/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
	usermocks "github.com/mathcale/go-api-boilerplate/internal/modules/user/mocks"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/usecases"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

type MeUseCaseTestSuite struct {
	suite.Suite
	gw *usermocks.UserGateway
}

func (s *MeUseCaseTestSuite) SetupTest() {
	s.gw = new(usermocks.UserGateway)
}

func TestMeUseCase(t *testing.T) {
	suite.Run(t, new(MeUseCaseTestSuite))
}

func (s *MeUseCaseTestSuite) TestSuccess() {
	s.Run("should return the user matching the given id", func() {
		userID := uuid.New()
		found := &domain.User{ID: userID, Email: "jane@example.com"}

		s.gw.On("GetUserByID", mock.Anything, userID).Return(found, nil)

		uc := usecases.NewMeUseCase(mocks.NoopLogger{}, s.gw)

		got, err := uc.Execute(context.Background(), userID)

		s.Require().NoError(err)
		s.Require().NotNil(got)
		s.Equal(userID, got.ID)
		s.gw.AssertExpectations(s.T())
	})
}

func (s *MeUseCaseTestSuite) TestNotFound() {
	s.Run("should return a not found business error when the user does not exist", func() {
		userID := uuid.New()

		s.gw.On("GetUserByID", mock.Anything, userID).Return(nil, nil)

		uc := usecases.NewMeUseCase(mocks.NoopLogger{}, s.gw)

		got, err := uc.Execute(context.Background(), userID)

		s.Require().Error(err)
		s.Nil(got)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.NotFoundKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1001_USER_NOT_FOUND, *appErr.BusinessCode())
	})
}

func (s *MeUseCaseTestSuite) TestGatewayError() {
	s.Run("should propagate the error returned by the gateway", func() {
		userID := uuid.New()
		gwErr := errors.New("db down")

		s.gw.On("GetUserByID", mock.Anything, userID).Return(nil, gwErr)

		uc := usecases.NewMeUseCase(mocks.NoopLogger{}, s.gw)

		got, err := uc.Execute(context.Background(), userID)

		s.Require().ErrorIs(err, gwErr)
		s.Nil(got)
	})
}
