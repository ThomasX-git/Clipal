package integration

import (
	"os"
	"strings"
	"testing"
)

func TestSmokeScriptBuildsUpstreamBinaryBeforeLaunch(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("../../scripts/smoke_test.sh")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	script := string(body)
	if strings.Contains(script, `go run "$tmpdir/upstream.go"`) {
		t.Fatalf("smoke script still starts upstream via go run")
	}
	if !strings.Contains(script, `go build -o "$tmpdir/upstream" "$tmpdir/upstream.go"`) {
		t.Fatalf("smoke script does not build upstream binary before launch")
	}
	if !strings.Contains(script, `"$tmpdir/upstream" >"$tmpdir/upstream.log" 2>&1 &`) {
		t.Fatalf("smoke script does not launch built upstream binary")
	}
}
