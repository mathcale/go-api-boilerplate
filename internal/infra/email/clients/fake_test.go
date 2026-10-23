package clients

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/tests/fixtures"
	"go-api-boilerplate/internal/tests/mocks"
)

type FakeClientTestSuite struct {
	suite.Suite
}

func TestFakeClient(t *testing.T) {
	suite.Run(t, new(FakeClientTestSuite))
}

func (s *FakeClientTestSuite) TestFake_Send() {
	s.Run("should log the email and succeed without actually sending it", func() {
		logger := new(mocks.Logger)
		logger.On("Info", mock.Anything, mock.Anything).Return()

		client := NewFake(logger, "Boilerplate", "no-reply@example.com")

		err := client.Send(fixtures.InfraEmailSendInput())

		s.Require().NoError(err)
		logger.AssertExpectations(s.T())
	})

	s.Run("should not log the email body", func() {
		input := fixtures.InfraEmailSendInput()
		var logged []any

		logger := new(mocks.Logger)
		logger.On("Info", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) { logged = args.Get(1).([]any) }).
			Return()

		client := NewFake(logger, "Boilerplate", "no-reply@example.com")

		s.Require().NoError(client.Send(input))
		s.Contains(logged, "recipient_email")
		s.Contains(logged, "subject")
		s.NotContains(logged, "body")
		s.NotContains(logged, input.HTMLBody)
	})
}
