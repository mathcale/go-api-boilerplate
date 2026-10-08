package clients

import (
	"github.com/mathcale/go-api-boilerplate/internal/core/email"
	"github.com/mathcale/go-api-boilerplate/internal/core/logger"
)

type fakeEmailClient struct {
	logger      logger.Logger
	senderName  string
	senderEmail string
}

func NewFakeEmailClient(l logger.Logger, senderName, senderEmail string) email.Client {
	return &fakeEmailClient{
		logger:      l,
		senderName:  senderName,
		senderEmail: senderEmail,
	}
}

func (c *fakeEmailClient) Send(input email.SendInput) error {
	c.logger.Info("[fake email client] email would be sent", map[string]interface{}{
		"from_name":          c.senderName,
		"from_email":         c.senderEmail,
		"recipient_name":     input.RecipientName,
		"recipient_email":    input.RecipientEmail,
		"subject":            input.Subject,
		"template_id":        input.TemplateID,
		"template_variables": input.TemplateVariables,
	})

	return nil
}
