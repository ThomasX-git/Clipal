package oauth

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/lansespirit/Clipal/internal/config"
)

func TestStoreSaveUses0600Permissions(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	cred := &Credential{
		Ref:          "codex-sean-example-com",
		Provider:     config.OAuthProviderCodex,
		Email:        "sean@example.com",
		AccountID:    "acct_123",
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Date(2026, 4, 18, 15, 0, 0, 0, time.UTC),
		LastRefresh:  time.Date(2026, 4, 18, 14, 30, 0, 0, time.UTC),
		Metadata: map[string]string{
			"id_token": "jwt-token",
		},
	}

	if err := store.Save(cred); err != nil {
		t.Fatalf("Save: %v", err)
	}

	path, err := store.resolvePath(config.OAuthProviderCodex, cred.Ref)
	if err != nil {
		t.Fatalf("resolvePath: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("perm = %o, want 0600", got)
	}
	if got := filepath.Base(path); got != "sean@example.com--codex-sean-example-com.json" {
		t.Fatalf("credential filename = %q", got)
	}
}

func TestCredentialFileName_PreservesReadableEmailPrefix(t *testing.T) {
	got := credentialFileName("eileenallen247719@hotmail.com", "7636ekkJJ[42")
	if got != "eileenallen247719@hotmail.com--7636ekkJJ[42.json" {
		t.Fatalf("credential filename = %q", got)
	}
}

func TestStoreLoadRoundTripPreservesCredentialMetadata(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	want := &Credential{
		Ref:          "codex-sean-example-com",
		Provider:     config.OAuthProviderCodex,
		Email:        "sean@example.com",
		AccountID:    "acct_123",
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Date(2026, 4, 18, 15, 0, 0, 0, time.UTC),
		LastRefresh:  time.Date(2026, 4, 18, 14, 30, 0, 0, time.UTC),
		Metadata: map[string]string{
			"id_token": "jwt-token",
		},
	}

	if err := store.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Load(config.OAuthProviderCodex, want.Ref)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("credential mismatch:\n got %#v\nwant %#v", got, want)
	}
}
