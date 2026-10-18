package email

type (
	SendInput struct {
		RecipientName  string
		RecipientEmail string
		Subject        string
		HTMLBody       string
	}

	Client interface {
		Send(input SendInput) error
	}
)
