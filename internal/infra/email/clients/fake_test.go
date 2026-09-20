package clients_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/infra/email"
	"github.com/mathcale/go-api-boilerplate/internal/infra/email/clients"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
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

		client := clients.NewFake(logger, "Boilerplate", "no-reply@example.com")

		err := client.Send(email.SendInput{
			RecipientName:  "Jane",
			RecipientEmail: "jane@example.com",
			Subject:        "Hello",
			TemplateID:     "welcome",
			TemplateVariables: map[string]interface{}{
				"first_name": "Jane",
			},
		})

		s.Require().NoError(err)
		logger.AssertExpectations(s.T())
	})
}
