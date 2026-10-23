package clientip

import (
	"fmt"
	"net/http"
	"net/netip"
	"strings"
)

const (
	forwardedForHeader = "X-Forwarded-For"
	ipv6KeyBits        = 64
)

type Resolver struct {
	trusted []netip.Prefix
}

func NewResolver(trusted []netip.Prefix) *Resolver {
	return &Resolver{trusted: trusted}
}

func ParseTrustedProxies(raw string) ([]netip.Prefix, error) {
	var prefixes []netip.Prefix

	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		prefix, err := parseEntry(entry)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy %q: %w", entry, err)
		}

		prefixes = append(prefixes, prefix)
	}

	return prefixes, nil
}

func (r *Resolver) Resolve(req *http.Request) string {
	remote, ok := parseAddr(req.RemoteAddr)
	if !ok {
		return req.RemoteAddr
	}

	if r.isTrusted(remote) {
		if client, found := r.fromForwardedFor(req); found {
			return key(client)
		}
	}

	return key(remote)
}

func (r *Resolver) fromForwardedFor(req *http.Request) (netip.Addr, bool) {
	var entries []string

	for _, value := range req.Header.Values(forwardedForHeader) {
		entries = append(entries, strings.Split(value, ",")...)
	}

	for i := len(entries) - 1; i >= 0; i-- {
		addr, ok := parseAddr(strings.TrimSpace(entries[i]))
		if !ok {
			return netip.Addr{}, false
		}

		if !r.isTrusted(addr) {
			return addr, true
		}
	}

	return netip.Addr{}, false
}

func (r *Resolver) isTrusted(addr netip.Addr) bool {
	for _, prefix := range r.trusted {
		if prefix.Contains(addr) {
			return true
		}
	}

	return false
}

func parseEntry(entry string) (netip.Prefix, error) {
	if strings.Contains(entry, "/") {
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			return netip.Prefix{}, err
		}

		return prefix.Masked(), nil
	}

	addr, err := netip.ParseAddr(entry)
	if err != nil {
		return netip.Prefix{}, err
	}

	addr = addr.Unmap().WithZone("")

	return netip.PrefixFrom(addr, addr.BitLen()), nil
}

func parseAddr(value string) (netip.Addr, bool) {
	if ap, err := netip.ParseAddrPort(value); err == nil {
		return ap.Addr().Unmap().WithZone(""), true
	}

	addr, err := netip.ParseAddr(strings.Trim(value, "[]"))
	if err != nil {
		return netip.Addr{}, false
	}

	return addr.Unmap().WithZone(""), true
}

func key(addr netip.Addr) string {
	if addr.Is6() {
		return netip.PrefixFrom(addr, ipv6KeyBits).Masked().Addr().String()
	}

	return addr.String()
}
