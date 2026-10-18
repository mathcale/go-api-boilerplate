package clients

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/mathcale/go-api-boilerplate/internal/infra/email"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

const (
	smtpDialTimeout = 10 * time.Second
	smtpSendTimeout = 30 * time.Second
)

var (
	ErrSMTPHeaderInjection = errors.New("email header value contains CR or LF")
	ErrSMTPInvalidAddress  = errors.New("invalid email address")
	ErrSMTPNoSTARTTLS      = errors.New("smtp server does not support STARTTLS")
)

type SMTPConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	ImplicitTLS bool
	SenderName  string
	SenderEmail string
}

type smtpClient struct {
	logger  logger.Logger
	cfg     SMTPConfig
	dial    func(network, addr string) (net.Conn, error)
	now     func() time.Time
	rootCAs *x509.CertPool
}

func NewSMTP(l logger.Logger, cfg SMTPConfig) email.Client {
	return &smtpClient{
		logger: l,
		cfg:    cfg,
		dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, smtpDialTimeout)
		},
		now: time.Now,
	}
}

func (c *smtpClient) Send(input email.SendInput) error {
	msg, err := c.buildMessage(input)
	if err != nil {
		return err
	}

	to, err := parseAddress(input.RecipientEmail)
	if err != nil {
		return err
	}

	from, err := parseAddress(c.cfg.SenderEmail)
	if err != nil {
		return err
	}

	if err := c.deliver(from, to, msg); err != nil {
		return fmt.Errorf("failed to send email via smtp: %w", err)
	}

	c.logger.Info("email sent", "recipient_email", input.RecipientEmail, "subject", input.Subject)

	return nil
}

func (c *smtpClient) deliver(from, to string, msg []byte) error {
	conn, err := c.dial("tcp", net.JoinHostPort(c.cfg.Host, strconv.Itoa(c.cfg.Port)))
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := c.handshake(conn)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	if c.cfg.Username != "" {
		auth := smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
	}

	if err = c.transmit(client, from, to, msg); err != nil {
		return err
	}

	return client.Quit()
}

func (c *smtpClient) handshake(conn net.Conn) (*smtp.Client, error) {
	tlsCfg := &tls.Config{
		ServerName: c.cfg.Host,
		MinVersion: tls.VersionTLS12,
		RootCAs:    c.rootCAs,
	}

	if err := conn.SetDeadline(time.Now().Add(smtpSendTimeout)); err != nil {
		return nil, fmt.Errorf("set deadline: %w", err)
	}

	if c.cfg.ImplicitTLS {
		tlsConn := tls.Client(conn, tlsCfg)
		if err := tlsConn.Handshake(); err != nil {
			return nil, fmt.Errorf("tls handshake: %w", err)
		}

		conn = tlsConn
	}

	client, err := smtp.NewClient(conn, c.cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("greeting: %w", err)
	}

	if c.cfg.ImplicitTLS {
		return client, nil
	}

	if ok, _ := client.Extension("STARTTLS"); !ok {
		_ = client.Close()

		return nil, ErrSMTPNoSTARTTLS
	}

	if err = client.StartTLS(tlsCfg); err != nil {
		_ = client.Close()

		return nil, fmt.Errorf("starttls: %w", err)
	}

	return client, nil
}

func (c *smtpClient) transmit(client *smtp.Client, from, to string, msg []byte) error {
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}

	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("write body: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("end data: %w", err)
	}

	return nil
}

func (c *smtpClient) buildMessage(input email.SendInput) ([]byte, error) {
	for _, v := range []string{
		input.RecipientName, input.RecipientEmail, input.Subject,
		c.cfg.SenderName, c.cfg.SenderEmail,
	} {
		if strings.ContainsAny(v, "\r\n") {
			return nil, ErrSMTPHeaderInjection
		}
	}

	from, err := parseAddress(c.cfg.SenderEmail)
	if err != nil {
		return nil, err
	}

	to, err := parseAddress(input.RecipientEmail)
	if err != nil {
		return nil, err
	}

	messageID, err := newMessageID(from)
	if err != nil {
		return nil, err
	}

	var b strings.Builder

	writeHeader := func(k, v string) { b.WriteString(k + ": " + v + "\r\n") }

	writeHeader("From", (&mail.Address{Name: c.cfg.SenderName, Address: from}).String())
	writeHeader("To", (&mail.Address{Name: input.RecipientName, Address: to}).String())
	writeHeader("Subject", mime.QEncoding.Encode("UTF-8", input.Subject))
	writeHeader("Date", c.now().Format(time.RFC1123Z))
	writeHeader("Message-ID", messageID)
	writeHeader("MIME-Version", "1.0")
	writeHeader("Content-Type", "text/html; charset=UTF-8")
	writeHeader("Content-Transfer-Encoding", "quoted-printable")
	b.WriteString("\r\n")

	qp := quotedprintable.NewWriter(&b)

	if _, err := qp.Write([]byte(input.HTMLBody)); err != nil {
		return nil, fmt.Errorf("failed to encode body: %w", err)
	}

	if err := qp.Close(); err != nil {
		return nil, fmt.Errorf("failed to encode body: %w", err)
	}

	return []byte(b.String()), nil
}

func parseAddress(raw string) (string, error) {
	if strings.ContainsAny(raw, "\r\n") {
		return "", ErrSMTPHeaderInjection
	}

	addr, err := mail.ParseAddress(raw)
	if err != nil || addr.Address != raw {
		return "", ErrSMTPInvalidAddress
	}

	return addr.Address, nil
}

func newMessageID(sender string) (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("failed to generate message id: %w", err)
	}

	domain := sender[strings.LastIndex(sender, "@")+1:]

	return "<" + hex.EncodeToString(raw) + "@" + domain + ">", nil
}
