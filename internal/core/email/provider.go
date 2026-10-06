package email

// SendInput is the transport-agnostic description of a transactional email.
// The TemplateID / TemplateVariables fields map cleanly onto templated
// providers (Mailjet, SendGrid, SES templates, ...). A concrete client decides
// how to render them.
type SendInput struct {
	RecipientName     string
	RecipientEmail    string
	Subject           string
	TemplateID        string
	TemplateVariables map[string]interface{}
}

// Client is the port every email provider implements. Swap the fake for a real
// SMTP/API client in the dependency injector without touching the use cases.
type Client interface {
	Send(input SendInput) error
}
