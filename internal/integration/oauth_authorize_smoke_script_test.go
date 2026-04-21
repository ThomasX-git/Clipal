package integration

import (
	"os"
	"strings"
	"testing"
)

func TestOAuthAuthorizeSmokeScriptExercisesAuthorizationBusinessRoute(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("../../scripts/oauth_authorize_smoke.sh")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	script := string(body)
	for _, want := range []string{
		`/api/oauth/providers/start`,
		`/api/oauth/sessions/$session_id`,
		`curl -fsS -L "$auth_url" >/dev/null`,
		`provider.get("oauth_provider", "")`,
		`provider.get("auth_type", "")`,
		`CLIPAL_OAUTH_CODEX_AUTH_URL=`,
		`CLIPAL_OAUTH_CODEX_TOKEN_URL=`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("oauth authorize smoke script missing %q", want)
		}
	}
}
