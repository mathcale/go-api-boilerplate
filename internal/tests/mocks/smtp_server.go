package mocks

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"time"
)

type SMTPServerOptions struct {
	ImplicitTLS     bool
	DisableSTARTTLS bool
	Username        string
	Password        string
}

type SMTPMessage struct {
	From string
	To   []string
	Data string
}

type SMTPServer struct {
	listener net.Listener
	tlsCfg   *tls.Config
	opts     SMTPServerOptions
	roots    *x509.CertPool
	wg       sync.WaitGroup

	mu       sync.Mutex
	messages []SMTPMessage
}

func NewSMTPServer(opts SMTPServerOptions) (*SMTPServer, error) {
	cert, pool, err := selfSignedCert("localhost")
	if err != nil {
		return nil, err
	}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	srv := &SMTPServer{
		listener: l,
		tlsCfg:   &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12},
		opts:     opts,
		roots:    pool,
	}

	srv.wg.Add(1)

	go srv.serve()

	return srv, nil
}

func (s *SMTPServer) Addr() string { return s.listener.Addr().String() }

func (s *SMTPServer) RootCAs() *x509.CertPool { return s.roots }

func (s *SMTPServer) Messages() []SMTPMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]SMTPMessage(nil), s.messages...)
}

func (s *SMTPServer) Close() {
	_ = s.listener.Close()
	s.wg.Wait()
}

func (s *SMTPServer) serve() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}

		s.wg.Add(1)

		go func() {
			defer s.wg.Done()
			s.handle(conn)
		}()
	}
}

type smtpSession struct {
	srv    *SMTPServer
	raw    net.Conn
	conn   net.Conn
	r      *bufio.Reader
	secure bool
	msg    SMTPMessage
}

func (s *SMTPServer) handle(raw net.Conn) {
	defer func() { _ = raw.Close() }()

	_ = raw.SetDeadline(time.Now().Add(10 * time.Second))

	sess := &smtpSession{srv: s, raw: raw, conn: raw}

	if s.opts.ImplicitTLS {
		sess.upgrade()
	} else {
		sess.r = bufio.NewReader(raw)
	}

	sess.reply("220 localhost ESMTP")

	for {
		line, err := sess.r.ReadString('\n')
		if err != nil || sess.command(strings.TrimRight(line, "\r\n")) {
			return
		}
	}
}

func (c *smtpSession) upgrade() {
	c.conn = tls.Server(c.raw, c.srv.tlsCfg)
	c.r = bufio.NewReader(c.conn)
	c.secure = true
}

func (c *smtpSession) reply(line string) {
	_, _ = fmt.Fprintf(c.conn, "%s\r\n", line)
}

func (c *smtpSession) command(line string) bool {
	cmd := strings.ToUpper(line)

	switch {
	case strings.HasPrefix(cmd, "EHLO"):
		c.ehlo()
	case cmd == "STARTTLS":
		c.reply("220 ready")
		c.upgrade()
	case strings.HasPrefix(cmd, "AUTH PLAIN "):
		if c.srv.authorized(line[len("AUTH PLAIN "):]) {
			c.reply("235 ok")
		} else {
			c.reply("535 authentication failed")
		}
	case strings.HasPrefix(cmd, "MAIL FROM:"):
		c.msg = SMTPMessage{From: extractAddress(line)}
		c.reply("250 ok")
	case strings.HasPrefix(cmd, "RCPT TO:"):
		c.msg.To = append(c.msg.To, extractAddress(line))
		c.reply("250 ok")
	case cmd == "DATA":
		c.reply("354 go ahead")
		c.msg.Data = readData(c.r)
		c.srv.store(c.msg)
		c.reply("250 queued")
	case cmd == "QUIT":
		c.reply("221 bye")

		return true
	default:
		c.reply("502 not implemented")
	}

	return false
}

func (c *smtpSession) ehlo() {
	c.reply("250-localhost")

	if !c.secure && !c.srv.opts.DisableSTARTTLS {
		c.reply("250-STARTTLS")
	}

	if c.secure && c.srv.opts.Username != "" {
		c.reply("250-AUTH PLAIN")
	}

	c.reply("250 8BITMIME")
}

func (s *SMTPServer) store(msg SMTPMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, msg)
}

func (s *SMTPServer) authorized(encoded string) bool {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return false
	}

	parts := strings.Split(string(decoded), "\x00")

	return len(parts) == 3 && parts[1] == s.opts.Username && parts[2] == s.opts.Password
}

func readData(r *bufio.Reader) string {
	var b strings.Builder

	for {
		line, err := r.ReadString('\n')
		if err != nil || line == ".\r\n" {
			return b.String()
		}

		b.WriteString(strings.TrimPrefix(line, "."))
	}
}

func extractAddress(line string) string {
	start := strings.Index(line, "<")
	end := strings.LastIndex(line, ">")

	if start < 0 || end < start {
		return ""
	}

	return line[start+1 : end]
}

func selfSignedCert(host string) (tls.Certificate, *x509.CertPool, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, nil, err
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: host},
		DNSNames:     []string{host},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, nil, err
	}

	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		return tls.Certificate{}, nil, err
	}

	pool := x509.NewCertPool()
	pool.AddCert(parsed)

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, pool, nil
}
