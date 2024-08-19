package counter

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

type CounterUseCaseTestSuite struct {
	suite.Suite
	LoggerMock *mocks.LoggerMock

	UseCase CounterUseCase
}

func (s *CounterUseCaseTestSuite) SetupTest() {
	s.LoggerMock = new(mocks.LoggerMock)

	s.UseCase = NewCounterUseCase(s.LoggerMock)
}

func (s *CounterUseCaseTestSuite) cleanMocks() {
	s.LoggerMock.ExpectedCalls = nil
	s.LoggerMock.Calls = nil
}

func TestCounterUseCase(t *testing.T) {
	suite.Run(t, new(CounterUseCaseTestSuite))
}

func (s *CounterUseCaseTestSuite) TestExecute() {
	s.Run("should increment counter", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Info", mock.Anything, mock.Anything).Return(nil)

		counter, err := s.UseCase.Execute()

		s.NoError(err)
		s.Equal(1, counter)
	})
}
