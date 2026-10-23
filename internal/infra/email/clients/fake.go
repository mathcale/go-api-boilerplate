package clients

import (
	"go-api-boilerplate/internal/infra/email"
	"go-api-boilerplate/internal/pkg/logger"
)

type fake struct {
	logger      logger.Logger
	senderName  string
	senderEmail string
}

func NewFake(l logger.Logger, senderName, senderEmail string) email.Client {
	return &fake{
		logger:      l,
		senderName:  senderName,
		senderEmail: senderEmail,
	}
}

func (c *fake) Send(input email.SendInput) error {
	c.logger.Info(
		"[fake email client] email would be sent",
		"from_name", c.senderName,
		"from_email", c.senderEmail,
		"recipient_name", input.RecipientName,
		"recipient_email", input.RecipientEmail,
		"subject", input.Subject,
	)

	return nil
}
