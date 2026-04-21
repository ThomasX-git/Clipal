package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lansespirit/Clipal/internal/config"
)

var (
	ErrCLIProxyAPINotCredential      = errors.New("not a cli-proxy-api oauth credential")
	ErrCLIProxyAPIUnsupportedType    = errors.New("unsupported cli-proxy-api oauth credential type")
	ErrCLIProxyAPIDisabledCredential = errors.New("cli-proxy-api oauth credential is disabled")
)

type cliProxyAPICredentialFile struct {
	Type         string `json:"type"`
	Email        string `json:"email"`
	AccountID    string `json:"account_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	Expired      string `json:"expired"`
	ExpiresAt    string `json:"expires_at"`
	LastRefresh  string `json:"last_refresh"`
	Disabled     bool   `json:"disabled"`
}

func ParseCLIProxyAPICredential(data []byte) (*Credential, error) {
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, ErrCLIProxyAPINotCredential
	}

	var raw cliProxyAPICredentialFile
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse cli-proxy-api credential json: %w", err)
	}

	provider := normalizeProvider(config.OAuthProvider(raw.Type))
	if provider == "" {
		return nil, ErrCLIProxyAPINotCredential
	}
	if raw.Disabled {
		return nil, ErrCLIProxyAPIDisabledCredential
	}

	switch provider {
	case config.OAuthProviderCodex:
		return parseCLIProxyAPICodexCredential(raw)
	default:
		return nil, fmt.Errorf("%w: %s", ErrCLIProxyAPIUnsupportedType, provider)
	}
}

func parseCLIProxyAPICodexCredential(raw cliProxyAPICredentialFile) (*Credential, error) {
	email := strings.TrimSpace(raw.Email)
	accountID := strings.TrimSpace(raw.AccountID)
	if tokenEmail, tokenAccountID := parseCodexIdentityToken(raw.IDToken); tokenEmail != "" || tokenAccountID != "" {
		if email == "" {
			email = tokenEmail
		}
		if accountID == "" {
			accountID = tokenAccountID
		}
	}

	if strings.TrimSpace(raw.AccessToken) == "" {
		return nil, fmt.Errorf("codex credential missing access_token")
	}
	if email == "" && accountID == "" {
		return nil, fmt.Errorf("codex credential missing email/account_id")
	}

	expiresAt, err := parseCLIProxyAPITime(firstNonEmpty(raw.Expired, raw.ExpiresAt))
	if err != nil {
		return nil, fmt.Errorf("parse codex credential expired time: %w", err)
	}
	lastRefresh, err := parseCLIProxyAPITime(raw.LastRefresh)
	if err != nil {
		return nil, fmt.Errorf("parse codex credential last_refresh time: %w", err)
	}

	cred := &Credential{
		Ref:          stableCredentialRef(config.OAuthProviderCodex, email, accountID),
		Provider:     config.OAuthProviderCodex,
		Email:        email,
		AccountID:    accountID,
		AccessToken:  strings.TrimSpace(raw.AccessToken),
		RefreshToken: strings.TrimSpace(raw.RefreshToken),
		ExpiresAt:    expiresAt,
		LastRefresh:  lastRefresh,
	}
	if idToken := strings.TrimSpace(raw.IDToken); idToken != "" {
		cred.Metadata = map[string]string{"id_token": idToken}
	}
	return cred, nil
}

func parseCLIProxyAPITime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339Nano, value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
