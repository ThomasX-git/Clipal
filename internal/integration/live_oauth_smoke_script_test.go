package integration

import (
	"os"
	"strings"
	"testing"
)

func TestLiveOAuthSmokeScriptUsesTemporaryCredentialCopyAndRefreshProbe(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("../../scripts/live_oauth_smoke.sh")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	script := string(body)
	for _, want := range []string{
		`--email EMAIL`,
		`default: gpt-5.4`,
		`CLIPAL_LIVE_OAUTH_EMAIL`,
		`list_credentials "$creds_json" "$OAUTH_EMAIL" "$OAUTH_REF" "$OAUTH_FILE"`,
		`cfgdir="$tmpdir/config"`,
		`oauth email not found`,
		`credential_path="$cfgdir/oauth/codex/$(basename "$oauth_source_path")"`,
		`auth_type: "oauth"`,
		`oauth_provider: "codex"`,
		`response.output_text.delta`,
		`"http://127.0.0.1:$clipal_port/clipal/v1/responses"`,
		`clipal-live-invalid-token`,
		`credential access_token was not replaced after refresh retry`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("live oauth smoke script missing %q", want)
		}
	}
}
