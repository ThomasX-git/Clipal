package config

import "testing"

func TestXXXCanonTrailingSlash(t *testing.T) {
	cases := []struct{ in, want string }{
		{"http://proxy:8080/", "http://proxy:8080/"},
		{"http://proxy:8080", "http://proxy:8080"},
		{"http://user:pass@proxy:8080", "http://user:pass@proxy:8080"},
		{"http://USER:PASS@PROXY:8080", "http://USER:PASS@proxy:8080"},
		{"http://proxy:8080/path", "http://proxy:8080/path"},
	}
	for _, c := range cases {
		got := CanonicalProxyURL(c.in)
		t.Logf("CanonicalProxyURL(%q) = %q (want %q, match=%v)", c.in, got, c.want, got == c.want)
	}
}
