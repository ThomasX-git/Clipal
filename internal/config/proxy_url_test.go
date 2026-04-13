package config

import (
	"testing"
)

func TestCanonicalProxyURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "whitespace only", input: "  ", want: ""},
		{name: "basic http", input: "http://proxy.example:8080", want: "http://proxy.example:8080"},
		{name: "trims whitespace", input: "  http://proxy.example:8080  ", want: "http://proxy.example:8080"},
		{name: "lowercases host", input: "http://PROXY.EXAMPLE:8080", want: "http://proxy.example:8080"},
		{name: "mixed case host", input: "http://Proxy.Example:8080", want: "http://proxy.example:8080"},
		{name: "removes default http port", input: "http://proxy.example:80", want: "http://proxy.example"},
		{name: "removes default https port", input: "https://proxy.example:443", want: "https://proxy.example"},
		{name: "removes default socks5 port", input: "socks5://proxy.example:1080", want: "socks5://proxy.example"},
		{name: "preserves non-default port", input: "http://proxy.example:8080", want: "http://proxy.example:8080"},
		{name: "scheme already lowercase", input: "http://proxy.example:8080", want: "http://proxy.example:8080"},
		{name: "socks5h", input: "socks5h://proxy.example:1080", want: "socks5h://proxy.example"},
		{name: "socks5h non-default port", input: "socks5h://proxy.example:9050", want: "socks5h://proxy.example:9050"},
		{name: "lowercases scheme", input: "HTTP://proxy.example:8080", want: "http://proxy.example:8080"},
		{name: "lowercases scheme and host together", input: "HTTPS://PROXY.EXAMPLE:443", want: "https://proxy.example"},
		{name: "invalid scheme returns trimmed input", input: "ftp://proxy.example:21", want: "ftp://proxy.example:21"},
		{name: "equivalent URLs produce identical canonical form", input: "  HTTP://PROXY.EXAMPLE:80  ", want: "http://proxy.example"},
		{name: "ip address host", input: "http://127.0.0.1:7890", want: "http://127.0.0.1:7890"},
		{name: "ipv6 with non-default port", input: "http://[::1]:8080", want: "http://[::1]:8080"},
		{name: "ipv6 strips default http port", input: "http://[::1]:80", want: "http://[::1]"},
		{name: "ipv6 strips default https port", input: "https://[::1]:443", want: "https://[::1]"},
		{name: "ipv6 without port", input: "http://[::1]", want: "http://[::1]"},
		{name: "ipv6 full address with default port", input: "http://[2001:db8::1]:80", want: "http://[2001:db8::1]"},
		{name: "ipv6 zone with non-default port preserves zone case", input: "http://[fe80::1%25En0]:8080", want: "http://[fe80::1%25En0]:8080"},
		{name: "ipv6 zone without port preserves zone case", input: "http://[fe80::1%25En0]", want: "http://[fe80::1%25En0]"},
		{name: "ipv6 zone strips default port without lowercasing zone", input: "http://[FE80::ABCD%25En0]:80", want: "http://[fe80::abcd%25En0]"},
		{name: "preserves userinfo case", input: "http://User:Pass@PROXY.EXAMPLE:8080", want: "http://User:Pass@proxy.example:8080"},
		{name: "userinfo with default port stripped", input: "http://user:pass@proxy.example:80", want: "http://user:pass@proxy.example"},
		{name: "preserves trailing slash", input: "http://proxy.example:8080/", want: "http://proxy.example:8080"},
		{name: "path rejected", input: "http://proxy.example:8080/path", want: "http://proxy.example:8080/path"},
		{name: "query rejected", input: "http://proxy.example:8080?x=1", want: "http://proxy.example:8080?x=1"},
		{name: "bare query rejected", input: "http://proxy.example:8080?", want: "http://proxy.example:8080?"},
		{name: "bare query after slash rejected", input: "http://proxy.example:8080/?", want: "http://proxy.example:8080/?"},
		{name: "fragment rejected", input: "http://proxy.example:8080#frag", want: "http://proxy.example:8080#frag"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CanonicalProxyURL(tt.input)
			if got != tt.want {
				t.Errorf("CanonicalProxyURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCanonicalProxyURL_Equivalence(t *testing.T) {
	t.Parallel()

	pairs := []struct{ a, b string }{
		{"http://PROXY.example:8080", "http://proxy.example:8080"},
		{"http://proxy.example:80", "http://proxy.example"},
		{"  http://proxy.example:8080  ", "http://proxy.example:8080"},
		{"HTTP://PROXY.EXAMPLE:80", "http://proxy.example"},
		{"http://user:pass@PROXY.EXAMPLE:80", "http://user:pass@proxy.example"},
		{"http://proxy.example/", "http://proxy.example"},
	}

	for _, p := range pairs {
		ca := CanonicalProxyURL(p.a)
		cb := CanonicalProxyURL(p.b)
		if ca != cb {
			t.Errorf("CanonicalProxyURL(%q) = %q, CanonicalProxyURL(%q) = %q; want equal", p.a, ca, p.b, cb)
		}
	}
}

func TestParseProxyURL_RejectsPathQueryFragment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "path", input: "http://proxy.example:8080/path"},
		{name: "query", input: "http://proxy.example:8080?x=1"},
		{name: "bare query", input: "http://proxy.example:8080?"},
		{name: "bare query after slash", input: "http://proxy.example:8080/?"},
		{name: "fragment", input: "http://proxy.example:8080#frag"},
		{name: "path and query", input: "http://proxy.example:8080/path?x=1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseProxyURL(tt.input)
			if err == nil {
				t.Errorf("ParseProxyURL(%q) succeeded; want rejection", tt.input)
			}
		})
	}
}

func TestParseProxyURL_AcceptsBareTrailingSlash(t *testing.T) {
	t.Parallel()

	parsed, err := ParseProxyURL("http://proxy.example:8080/")
	if err != nil {
		t.Fatalf("ParseProxyURL with trailing slash: %v", err)
	}
	if parsed.Path != "" {
		t.Errorf("Path = %q, want empty (trailing slash should be stripped)", parsed.Path)
	}
}

func TestProvider_CanonicalProxyURL(t *testing.T) {
	t.Parallel()

	p := Provider{ProxyURL: "  HTTP://PROXY.EXAMPLE:80  "}
	if got := p.CanonicalProxyURL(); got != "http://proxy.example" {
		t.Errorf("Provider.CanonicalProxyURL() = %q, want %q", got, "http://proxy.example")
	}
	// Normalized accessor preserves original text
	if got := p.NormalizedProxyURL(); got != "HTTP://PROXY.EXAMPLE:80" {
		t.Errorf("Provider.NormalizedProxyURL() = %q, want %q", got, "HTTP://PROXY.EXAMPLE:80")
	}
}

func TestGlobalConfig_CanonicalUpstreamProxyURL(t *testing.T) {
	t.Parallel()

	g := GlobalConfig{UpstreamProxyURL: "  HTTP://PROXY.EXAMPLE:80  "}
	if got := g.CanonicalUpstreamProxyURL(); got != "http://proxy.example" {
		t.Errorf("GlobalConfig.CanonicalUpstreamProxyURL() = %q, want %q", got, "http://proxy.example")
	}
	// Normalized accessor preserves original text
	if got := g.NormalizedUpstreamProxyURL(); got != "HTTP://PROXY.EXAMPLE:80" {
		t.Errorf("GlobalConfig.NormalizedUpstreamProxyURL() = %q, want %q", got, "HTTP://PROXY.EXAMPLE:80")
	}
}
