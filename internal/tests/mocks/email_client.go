package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/go-api-boilerplate/internal/infra/email"
)

type EmailClient struct {
	mock.Mock
}

func (m *EmailClient) Send(input email.SendInput) error {
	return m.Called(input).Error(0)
}
