package oauth

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lansespirit/Clipal/internal/config"
)

func TestSessionPollExpiresAutomatically(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 4, 18, 16, 0, 0, 0, time.UTC)
	current := now

	svc := NewService(dir,
		WithNowFunc(func() time.Time { return current }),
		WithSessionTTL(30*time.Second),
		WithCodexClient(&CodexClient{
			AuthURL:      "https://auth.openai.com/oauth/authorize",
			TokenURL:     "https://auth.openai.com/oauth/token",
			ClientID:     "test-client",
			CallbackHost: "127.0.0.1",
			CallbackPort: 0,
			CallbackPath: "/auth/callback",
			Now:          func() time.Time { return current },
		}),
	)

	session, err := svc.StartLogin(config.OAuthProviderCodex)
	if err != nil {
		t.Fatalf("StartLogin: %v", err)
	}

	current = current.Add(45 * time.Second)
	got, err := svc.PollLogin(session.ID)
	if err != nil {
		t.Fatalf("PollLogin: %v", err)
	}
	if got.Status != LoginStatusExpired {
		t.Fatalf("status = %q, want %q", got.Status, LoginStatusExpired)
	}
}

func TestRefreshIfNeededCoalescesConcurrentCallers(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 4, 18, 18, 0, 0, 0, time.UTC)
	var refreshCalls int32

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("grant_type"); got != "refresh_token" {
			t.Fatalf("grant_type = %q, want refresh_token", got)
		}
		atomic.AddInt32(&refreshCalls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, fmt.Sprintf(`{"access_token":"access-2","refresh_token":"refresh-2","id_token":"%s","expires_in":3600}`, testJWT("sean@example.com", "acct_123")))
	}))
	defer tokenServer.Close()

	store := NewStore(dir)
	if err := store.Save(&Credential{
		Ref:          "codex-sean-example-com",
		Provider:     config.OAuthProviderCodex,
		Email:        "sean@example.com",
		AccountID:    "acct_123",
		AccessToken:  "access-1",
		RefreshToken: "refresh-1",
		ExpiresAt:    now.Add(10 * time.Second),
		LastRefresh:  now.Add(-time.Hour),
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	svc := NewService(dir,
		WithNowFunc(func() time.Time { return now }),
		WithRefreshSkew(30*time.Second),
		WithCodexClient(&CodexClient{
			AuthURL:      "https://auth.openai.com/oauth/authorize",
			TokenURL:     tokenServer.URL,
			ClientID:     "test-client",
			CallbackHost: "127.0.0.1",
			CallbackPort: 0,
			CallbackPath: "/auth/callback",
			HTTPClient:   tokenServer.Client(),
			Now:          func() time.Time { return now },
		}),
	)

	const callers = 6
	var wg sync.WaitGroup
	results := make(chan *Credential, callers)
	errors := make(chan error, callers)
	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cred, err := svc.RefreshIfNeeded(context.Background(), config.OAuthProviderCodex, "codex-sean-example-com")
			if err != nil {
				errors <- err
				return
			}
			results <- cred
		}()
	}
	wg.Wait()
	close(results)
	close(errors)

	for err := range errors {
		t.Fatalf("RefreshIfNeeded: %v", err)
	}
	if got := atomic.LoadInt32(&refreshCalls); got != 1 {
		t.Fatalf("refresh calls = %d, want 1", got)
	}
	for cred := range results {
		if cred.AccessToken != "access-2" {
			t.Fatalf("access token = %q, want access-2", cred.AccessToken)
		}
		if cred.RefreshToken != "refresh-2" {
			t.Fatalf("refresh token = %q, want refresh-2", cred.RefreshToken)
		}
	}

	loaded, err := svc.Load(config.OAuthProviderCodex, "codex-sean-example-com")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.AccessToken != "access-2" {
		t.Fatalf("stored access token = %q, want access-2", loaded.AccessToken)
	}
}

func testJWT(email string, accountID string) string {
	header := `{"alg":"none","typ":"JWT"}`
	payload := fmt.Sprintf(`{"email":"%s","sub":"sub_123","https://api.openai.com/auth":{"chatgpt_account_id":"%s"}}`, email, accountID)
	return encodeSegment(header) + "." + encodeSegment(payload) + "."
}

func encodeSegment(v string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(v))
}
