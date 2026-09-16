package clients

import (
	"github.com/mathcale/go-api-boilerplate/internal/infra/email"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

// fake is the default email client. Instead of delivering messages it logs the
// full SendInput, so flows that depend on email (account confirmation, password
// recovery) work out-of-the-box in development without any external provider.
// Replace it with a real client (SMTP, Mailjet, SES, ...) in the injector.
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
