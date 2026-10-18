package clientip

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClientIPTestSuite struct {
	suite.Suite
}

func TestClientIP(t *testing.T) {
	suite.Run(t, new(ClientIPTestSuite))
}

func (s *ClientIPTestSuite) newRequest(remote string, xff ...string) *http.Request {
	s.T().Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remote

	for _, v := range xff {
		req.Header.Add(forwardedForHeader, v)
	}

	return req
}

func (s *ClientIPTestSuite) resolver(raw string) *Resolver {
	s.T().Helper()

	prefixes, err := ParseTrustedProxies(raw)
	s.Require().NoError(err)

	return NewResolver(prefixes)
}

func (s *ClientIPTestSuite) TestParseTrustedProxies() {
	s.Run("should return no prefixes for an empty value", func() {
		prefixes, err := ParseTrustedProxies("")

		s.Require().NoError(err)
		s.Empty(prefixes)
	})

	s.Run("should parse IPs and CIDRs ignoring whitespace and blanks", func() {
		prefixes, err := ParseTrustedProxies(" 10.0.0.1 , 192.168.0.0/16,, ::1 ")

		s.Require().NoError(err)
		s.Len(prefixes, 3)
		s.Equal("10.0.0.1/32", prefixes[0].String())
		s.Equal("192.168.0.0/16", prefixes[1].String())
		s.Equal("::1/128", prefixes[2].String())
	})

	s.Run("should reject an invalid IP", func() {
		_, err := ParseTrustedProxies("10.0.0.1,nope")

		s.ErrorContains(err, "nope")
	})

	s.Run("should reject an invalid CIDR", func() {
		_, err := ParseTrustedProxies("10.0.0.0/33")

		s.ErrorContains(err, "10.0.0.0/33")
	})
}

func (s *ClientIPTestSuite) TestResolve() {
	s.Run("should use the remote address when no proxies are trusted", func() {
		req := s.newRequest("203.0.113.9:4321", "198.51.100.7")

		s.Equal("203.0.113.9", s.resolver("").Resolve(req))
	})

	s.Run("should take the first untrusted entry from the right of the chain", func() {
		req := s.newRequest("10.0.0.1:80", "1.1.1.1, 198.51.100.7, 10.0.0.2")

		s.Equal("198.51.100.7", s.resolver("10.0.0.0/8").Resolve(req))
	})

	s.Run("should read a chain split across several header lines", func() {
		req := s.newRequest("10.0.0.1:80", "1.1.1.1", "198.51.100.7, 10.0.0.2")

		s.Equal("198.51.100.7", s.resolver("10.0.0.0/8").Resolve(req))
	})

	s.Run("should ignore the header when the remote is not trusted", func() {
		req := s.newRequest("203.0.113.9:4321", "198.51.100.7")

		s.Equal("203.0.113.9", s.resolver("10.0.0.0/8").Resolve(req))
	})

	s.Run("should use the remote address when the header is absent", func() {
		req := s.newRequest("10.0.0.1:80")

		s.Equal("10.0.0.1", s.resolver("10.0.0.0/8").Resolve(req))
	})

	s.Run("should use the remote address when every entry is trusted", func() {
		req := s.newRequest("10.0.0.1:80", "10.0.0.5, 10.0.0.6")

		s.Equal("10.0.0.1", s.resolver("10.0.0.0/8").Resolve(req))
	})

	s.Run("should accept entries carrying a port", func() {
		req := s.newRequest("10.0.0.1:80", "198.51.100.7:5555")

		s.Equal("198.51.100.7", s.resolver("10.0.0.0/8").Resolve(req))
	})

	s.Run("should key IPv6 addresses by their /64", func() {
		a := s.newRequest("[2001:db8:1:2:aaaa::1]:80")
		b := s.newRequest("[2001:db8:1:2:bbbb::9]:80")
		other := s.newRequest("[2001:db8:1:3::1]:80")
		r := s.resolver("")

		s.Equal("2001:db8:1:2::", r.Resolve(a))
		s.Equal(r.Resolve(a), r.Resolve(b))
		s.NotEqual(r.Resolve(a), r.Resolve(other))
	})

	s.Run("should key a forwarded IPv6 client by its /64", func() {
		req := s.newRequest("10.0.0.1:80", "2001:db8:1:2:cccc::5")

		s.Equal("2001:db8:1:2::", s.resolver("10.0.0.1").Resolve(req))
	})

	s.Run("should unmap IPv4-mapped IPv6 addresses", func() {
		req := s.newRequest("[::ffff:203.0.113.9]:80")

		s.Equal("203.0.113.9", s.resolver("").Resolve(req))
	})

	s.Run("should fall back to the remote address on a malformed entry", func() {
		req := s.newRequest("10.0.0.1:80", "198.51.100.7, garbage, 10.0.0.2")

		s.Equal("10.0.0.1", s.resolver("10.0.0.0/8").Resolve(req))
	})

	s.Run("should return the raw remote address when it cannot be parsed", func() {
		req := s.newRequest("not-an-address")

		s.Equal("not-an-address", s.resolver("10.0.0.0/8").Resolve(req))
	})
}
