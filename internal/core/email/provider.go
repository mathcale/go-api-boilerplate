package email

type (
	SendInput struct {
		RecipientName     string
		RecipientEmail    string
		Subject           string
		TemplateID        string
		TemplateVariables map[string]interface{}
	}

	Client interface {
		Send(input SendInput) error
	}
)
