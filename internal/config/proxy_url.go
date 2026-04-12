package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

const (
	supportedProxyURLSchemeList = "http, https, socks5, or socks5h"
	supportedProxyURLPrefixList = "http://, https://, socks5://, or socks5h://"
)

// ParseProxyURL validates a configured proxy URL and normalizes its scheme.
func ParseProxyURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !parsed.IsAbs() || parsed.Host == "" {
		return nil, fmt.Errorf("proxy_url must be an absolute %s URL", supportedProxyURLPrefixList)
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	switch parsed.Scheme {
	case "http", "https", "socks5", "socks5h":
		return parsed, nil
	default:
		return nil, fmt.Errorf("proxy_url scheme must be %s", supportedProxyURLSchemeList)
	}
}

// CanonicalProxyURL returns a canonical form of the given proxy URL.
// It trims whitespace, parses and validates the URL via ParseProxyURL, and
// normalizes the result so that semantically equivalent URLs
// (e.g. differing only in host case or default port) produce identical strings.
// If the URL is empty or cannot be parsed, the trimmed input is returned unchanged.
func CanonicalProxyURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	parsed, err := ParseProxyURL(trimmed)
	if err != nil {
		// Unparseable URLs are returned as-is. This is safe because
		// ParseProxyURL would also reject them when building the HTTP client,
		// so a malformed URL can never actually be used as a proxy.
		return trimmed
	}
	parsed.Host = canonicalHost(parsed.Host, parsed.Scheme)
	return parsed.String()
}

// canonicalHost lowercases the hostname and strips the port when it matches
// the default port for the given scheme.
func canonicalHost(host, scheme string) string {
	hostname, port, err := net.SplitHostPort(host)
	if err != nil {
		return canonicalHostWithoutPort(host)
	}
	hostname = canonicalHostname(hostname)
	defaultPort := defaultPortForScheme(scheme)
	if port == defaultPort {
		if strings.Contains(hostname, ":") {
			return "[" + hostname + "]"
		}
		return hostname
	}
	return net.JoinHostPort(hostname, port)
}

// canonicalHostWithoutPort handles bare hosts (no port) from net.SplitHostPort
// failure. For bracketed IPv6 literals like "[::1]", it strips the brackets,
// lowercases the address, and re-wraps so the result stays valid.
func canonicalHostWithoutPort(host string) string {
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return "[" + canonicalHostname(host[1:len(host)-1]) + "]"
	}
	return canonicalHostname(host)
}

// canonicalHostname lowercases the address or DNS name while preserving the
// case of any IPv6 zone identifier, which is interface-name text.
func canonicalHostname(hostname string) string {
	if zoneSep := strings.Index(hostname, "%"); zoneSep >= 0 {
		return strings.ToLower(hostname[:zoneSep]) + hostname[zoneSep:]
	}
	return strings.ToLower(hostname)
}

// defaultPortForScheme returns the conventional default port for a proxy scheme.
// SOCKS proxies default to 1080; HTTP/HTTPS fall back to their standard ports.
func defaultPortForScheme(scheme string) string {
	switch scheme {
	case "http":
		return "80"
	case "https":
		return "443"
	case "socks5", "socks5h":
		return "1080"
	default:
		return ""
	}
}

// ValidateProxyURL reports whether a configured proxy URL is supported.
func ValidateProxyURL(raw string) error {
	_, err := ParseProxyURL(raw)
	return err
}
