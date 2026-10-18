package clients

import (
	"errors"
	"io"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/infra/email"
	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

type SMTPClientTestSuite struct {
	suite.Suite
	fixedNow time.Time
}

func TestSMTPClient(t *testing.T) {
	suite.Run(t, new(SMTPClientTestSuite))
}

func (s *SMTPClientTestSuite) SetupTest() {
	s.fixedNow = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
}

func (s *SMTPClientTestSuite) newClient(cfg SMTPConfig) *smtpClient {
	if cfg.SenderEmail == "" {
		cfg.SenderEmail = "no-reply@example.com"
	}

	c := NewSMTP(mocks.NoopLogger{}, cfg).(*smtpClient)
	c.now = func() time.Time { return s.fixedNow }

	return c
}

func (s *SMTPClientTestSuite) startServer(
	opts mocks.SMTPServerOptions,
	cfg SMTPConfig,
) (*mocks.SMTPServer, *smtpClient) {
	s.T().Helper()

	srv, err := mocks.NewSMTPServer(opts)
	s.Require().NoError(err)
	s.T().Cleanup(srv.Close)

	host, port, err := net.SplitHostPort(srv.Addr())
	s.Require().NoError(err)

	p, err := strconv.Atoi(port)
	s.Require().NoError(err)

	cfg.Host = "localhost"
	cfg.Port = p
	cfg.ImplicitTLS = opts.ImplicitTLS

	c := s.newClient(cfg)
	c.rootCAs = srv.RootCAs()
	c.dial = func(network, _ string) (net.Conn, error) {
		return net.DialTimeout(network, net.JoinHostPort(host, port), time.Second)
	}

	return srv, c
}

func (s *SMTPClientTestSuite) parse(raw []byte) (*mail.Message, string) {
	s.T().Helper()

	msg, err := mail.ReadMessage(strings.NewReader(string(raw)))
	s.Require().NoError(err)

	body, err := io.ReadAll(quotedprintable.NewReader(msg.Body))
	s.Require().NoError(err)

	return msg, string(body)
}

func (s *SMTPClientTestSuite) TestSMTP_BuildMessage() {
	s.Run("should build a quoted-printable html message with all headers", func() {
		c := s.newClient(SMTPConfig{SenderName: "Boilerplate"})
		input := fixtures.InfraEmailSendInput()
		input.HTMLBody = "<p>Olá, café</p>"

		raw, err := c.buildMessage(input)

		s.Require().NoError(err)
		msg, body := s.parse(raw)
		s.Equal("<p>Olá, café</p>", body)
		s.Equal("text/html; charset=UTF-8", msg.Header.Get("Content-Type"))
		s.Equal("quoted-printable", msg.Header.Get("Content-Transfer-Encoding"))
		s.Equal("1.0", msg.Header.Get("MIME-Version"))
		s.Equal("Hello", msg.Header.Get("Subject"))
		s.Contains(msg.Header.Get("Message-ID"), "@example.com>")

		date, err := msg.Header.Date()
		s.Require().NoError(err)
		s.True(s.fixedNow.Equal(date))

		from, err := msg.Header.AddressList("From")
		s.Require().NoError(err)
		s.Equal("Boilerplate", from[0].Name)
		s.Equal("no-reply@example.com", from[0].Address)

		to, err := msg.Header.AddressList("To")
		s.Require().NoError(err)
		s.Equal("Jane", to[0].Name)
		s.Equal(fixtures.InfraUserEmail, to[0].Address)
		s.Contains(string(raw), "=C3=A1")
	})

	s.Run("should q-encode non-ascii names and subjects", func() {
		c := s.newClient(SMTPConfig{SenderName: "Equipe Ação"})
		input := fixtures.InfraSMTPSendInputWithName("José, Jr.")
		input.Subject = "Confirmação de conta"

		raw, err := c.buildMessage(input)

		s.Require().NoError(err)
		msg, _ := s.parse(raw)
		s.Contains(msg.Header.Get("Subject"), "=?")
		s.NotContains(string(raw[:strings.Index(string(raw), "\r\n\r\n")]), "ç")

		subject, err := new(mime.WordDecoder).DecodeHeader(msg.Header.Get("Subject"))
		s.Require().NoError(err)
		s.Equal("Confirmação de conta", subject)

		to, err := msg.Header.AddressList("To")
		s.Require().NoError(err)
		s.Equal("José, Jr.", to[0].Name)

		from, err := msg.Header.AddressList("From")
		s.Require().NoError(err)
		s.Equal("Equipe Ação", from[0].Name)
	})

	s.Run("should generate a unique message id per message", func() {
		c := s.newClient(SMTPConfig{})

		a, err := c.buildMessage(fixtures.InfraEmailSendInput())
		s.Require().NoError(err)
		b, err := c.buildMessage(fixtures.InfraEmailSendInput())
		s.Require().NoError(err)

		ma, _ := s.parse(a)
		mb, _ := s.parse(b)
		s.NotEqual(ma.Header.Get("Message-ID"), mb.Header.Get("Message-ID"))
	})

	s.Run("should reject CR or LF in header values", func() {
		cases := map[string]func(in *email.SendInput, cfg *SMTPConfig){
			"recipient name": func(in *email.SendInput, _ *SMTPConfig) {
				in.RecipientName = "a\r\nBcc: x@y.z"
			},
			"recipient email": func(in *email.SendInput, _ *SMTPConfig) {
				in.RecipientEmail = "a@b.c\nBcc: x@y.z"
			},
			"subject": func(in *email.SendInput, _ *SMTPConfig) {
				in.Subject = "Hi\r\nBcc: x@y.z"
			},
			"sender name": func(_ *email.SendInput, cfg *SMTPConfig) {
				cfg.SenderName = "x\nBcc: x@y.z"
			},
		}

		for name, mutate := range cases {
			input := fixtures.InfraEmailSendInput()
			cfg := SMTPConfig{}
			mutate(&input, &cfg)

			_, err := s.newClient(cfg).buildMessage(input)

			s.ErrorIs(err, ErrSMTPHeaderInjection, name)
		}
	})

	s.Run("should reject invalid recipient addresses", func() {
		for _, addr := range []string{"", "not-an-email", "Jane <jane@example.com>"} {
			input := fixtures.InfraEmailSendInput()
			input.RecipientEmail = addr

			_, err := s.newClient(SMTPConfig{}).buildMessage(input)

			s.ErrorIs(err, ErrSMTPInvalidAddress, addr)
		}
	})
}

func (s *SMTPClientTestSuite) TestSMTP_Send() {
	s.Run("should deliver over STARTTLS with PLAIN auth", func() {
		srv, c := s.startServer(
			mocks.SMTPServerOptions{Username: "mailer", Password: fixtures.ConfigSMTPCredential},
			SMTPConfig{Username: "mailer", Password: fixtures.ConfigSMTPCredential},
		)

		err := c.Send(fixtures.InfraEmailSendInput())

		s.Require().NoError(err)
		msgs := srv.Messages()
		s.Require().Len(msgs, 1)
		s.Equal("no-reply@example.com", msgs[0].From)
		s.Equal([]string{fixtures.InfraUserEmail}, msgs[0].To)
		_, body := s.parse([]byte(msgs[0].Data))
		s.Equal("<p>Hello Jane</p>", strings.TrimSpace(body))
	})

	s.Run("should deliver over implicit TLS without auth", func() {
		srv, c := s.startServer(mocks.SMTPServerOptions{ImplicitTLS: true}, SMTPConfig{})

		err := c.Send(fixtures.InfraEmailSendInput())

		s.Require().NoError(err)
		s.Len(srv.Messages(), 1)
	})

	s.Run("should fail when the server does not support STARTTLS", func() {
		srv, c := s.startServer(mocks.SMTPServerOptions{DisableSTARTTLS: true}, SMTPConfig{})

		err := c.Send(fixtures.InfraEmailSendInput())

		s.Require().ErrorIs(err, ErrSMTPNoSTARTTLS)
		s.Empty(srv.Messages())
	})

	s.Run("should fail without leaking the password when auth is rejected", func() {
		srv, c := s.startServer(
			mocks.SMTPServerOptions{Username: "mailer", Password: "right"},
			SMTPConfig{Username: "mailer", Password: fixtures.ConfigSMTPCredential},
		)

		err := c.Send(fixtures.InfraEmailSendInput())

		s.Require().Error(err)
		s.Contains(err.Error(), "auth")
		s.NotContains(err.Error(), fixtures.ConfigSMTPCredential)
		s.Empty(srv.Messages())
	})

	s.Run("should fail when the server certificate is not trusted", func() {
		srv, c := s.startServer(mocks.SMTPServerOptions{}, SMTPConfig{})
		c.rootCAs = nil

		err := c.Send(fixtures.InfraEmailSendInput())

		s.Require().Error(err)
		s.Empty(srv.Messages())
	})

	s.Run("should fail when the connection cannot be established", func() {
		c := s.newClient(SMTPConfig{Host: "localhost", Port: 25})
		c.dial = func(string, string) (net.Conn, error) { return nil, errors.New("refused") }

		err := c.Send(fixtures.InfraEmailSendInput())

		s.Require().Error(err)
		s.Contains(err.Error(), "refused")
	})

	s.Run("should not dial when the message is invalid", func() {
		c := s.newClient(SMTPConfig{})
		c.dial = func(string, string) (net.Conn, error) {
			s.Fail("dial must not be called")

			return nil, errors.New("unexpected")
		}
		input := fixtures.InfraEmailSendInput()
		input.Subject = "x\r\nBcc: a@b.c"

		s.Require().ErrorIs(c.Send(input), ErrSMTPHeaderInjection)
	})
}
