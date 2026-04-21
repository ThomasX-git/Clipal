package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lansespirit/Clipal/internal/config"
)

type Option func(*Service)

type Service struct {
	store       *Store
	codex       *CodexClient
	now         func() time.Time
	sessionTTL  time.Duration
	refreshSkew time.Duration

	mu        sync.Mutex
	sessions  map[string]*LoginSession
	refreshes map[string]*refreshCall
}

type refreshCall struct {
	done chan struct{}
	cred *Credential
	err  error
}

func NewService(configDir string, opts ...Option) *Service {
	svc := &Service{
		store:       NewStore(configDir),
		codex:       NewCodexClient(),
		now:         time.Now,
		sessionTTL:  5 * time.Minute,
		refreshSkew: 30 * time.Second,
		sessions:    make(map[string]*LoginSession),
		refreshes:   make(map[string]*refreshCall),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(svc)
		}
	}
	return svc
}

func WithNowFunc(fn func() time.Time) Option {
	return func(s *Service) {
		if fn != nil {
			s.now = fn
		}
	}
}

func WithCodexClient(client *CodexClient) Option {
	return func(s *Service) {
		if client != nil {
			s.codex = client
		}
	}
}

func WithSessionTTL(ttl time.Duration) Option {
	return func(s *Service) {
		if ttl > 0 {
			s.sessionTTL = ttl
		}
	}
}

func WithRefreshSkew(skew time.Duration) Option {
	return func(s *Service) {
		if skew >= 0 {
			s.refreshSkew = skew
		}
	}
}

func (s *Service) StartLogin(provider config.OAuthProvider) (*LoginSession, error) {
	provider = normalizeProvider(provider)
	if provider != config.OAuthProviderCodex {
		return nil, fmt.Errorf("unsupported oauth provider %q", provider)
	}

	pkce, err := GeneratePKCECodes()
	if err != nil {
		return nil, err
	}
	callback, redirectURI, err := startCallbackServer(s.codex.callbackHost(), s.codex.callbackPort(), s.codex.callbackPath())
	if err != nil {
		return nil, err
	}

	sessionID, err := randomID()
	if err != nil {
		_ = callback.Close()
		return nil, err
	}
	authURL, err := s.codex.GenerateAuthURL(sessionID, redirectURI, pkce)
	if err != nil {
		_ = callback.Close()
		return nil, err
	}

	session := &LoginSession{
		ID:          sessionID,
		Provider:    provider,
		AuthURL:     authURL,
		Status:      LoginStatusPending,
		ExpiresAt:   s.now().Add(s.sessionTTL),
		pkce:        pkce,
		redirectURI: redirectURI,
		callback:    callback,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweepExpiredSessionsLocked()
	s.sessions[session.ID] = session
	return session.Clone(), nil
}

func (s *Service) PollLogin(sessionID string) (*LoginSession, error) {
	s.mu.Lock()
	session, ok := s.sessions[strings.TrimSpace(sessionID)]
	if !ok {
		s.mu.Unlock()
		return nil, fmt.Errorf("oauth session not found")
	}
	if session.Status == LoginStatusPending && !session.ExpiresAt.IsZero() && !session.ExpiresAt.After(s.now()) {
		callback := session.callback
		session.callback = nil
		session.Status = LoginStatusExpired
		session.Error = "oauth session expired"
		s.mu.Unlock()
		if callback != nil {
			_ = callback.Close()
		}
		return session.Clone(), nil
	}
	if session.Status != LoginStatusPending || session.callback == nil {
		out := session.Clone()
		s.mu.Unlock()
		return out, nil
	}
	callback := session.callback
	redirectURI := session.redirectURI
	pkce := session.pkce
	provider := session.Provider
	expectedState := session.ID
	s.mu.Unlock()

	result, ok := callback.Poll()
	if !ok {
		s.mu.Lock()
		out := s.sessions[sessionID].Clone()
		s.mu.Unlock()
		return out, nil
	}
	if result.Error != "" {
		return s.finishSessionWithError(sessionID, result.Error, callback), nil
	}
	if result.State != expectedState {
		return s.finishSessionWithError(sessionID, "oauth state mismatch", callback), nil
	}

	var cred *Credential
	var err error
	switch provider {
	case config.OAuthProviderCodex:
		cred, err = s.codex.ExchangeCode(context.Background(), result.Code, redirectURI, pkce)
	default:
		err = fmt.Errorf("unsupported oauth provider %q", provider)
	}
	if err != nil {
		return s.finishSessionWithError(sessionID, err.Error(), callback), nil
	}
	if err := s.store.Save(cred); err != nil {
		return s.finishSessionWithError(sessionID, err.Error(), callback), nil
	}

	_ = callback.Close()
	s.mu.Lock()
	defer s.mu.Unlock()
	session = s.sessions[sessionID]
	session.callback = nil
	session.Status = LoginStatusCompleted
	session.CredentialRef = cred.Ref
	session.Email = cred.Email
	session.Error = ""
	return session.Clone(), nil
}

func (s *Service) Load(provider config.OAuthProvider, ref string) (*Credential, error) {
	return s.store.Load(provider, ref)
}

func (s *Service) Store() *Store {
	return s.store
}

func (s *Service) List(provider config.OAuthProvider) ([]Credential, error) {
	return s.store.List(provider)
}

func (s *Service) Delete(provider config.OAuthProvider, ref string) error {
	return s.store.Delete(provider, ref)
}

func (s *Service) RefreshIfNeeded(ctx context.Context, provider config.OAuthProvider, ref string) (*Credential, error) {
	return s.refresh(ctx, provider, ref, false)
}

func (s *Service) Refresh(ctx context.Context, provider config.OAuthProvider, ref string) (*Credential, error) {
	return s.refresh(ctx, provider, ref, true)
}

func (s *Service) refresh(ctx context.Context, provider config.OAuthProvider, ref string, force bool) (*Credential, error) {
	cred, err := s.store.Load(provider, ref)
	if err != nil {
		return nil, err
	}
	if !force && (!cred.NeedsRefresh(s.now(), s.refreshSkew) || strings.TrimSpace(cred.RefreshToken) == "") {
		return cred, nil
	}
	if strings.TrimSpace(cred.RefreshToken) == "" {
		return nil, fmt.Errorf("oauth credential %q has no refresh token", strings.TrimSpace(ref))
	}

	key := string(cred.Provider) + ":" + cred.Ref
	s.mu.Lock()
	if call, ok := s.refreshes[key]; ok {
		s.mu.Unlock()
		<-call.done
		return call.cred.Clone(), call.err
	}
	call := &refreshCall{done: make(chan struct{})}
	s.refreshes[key] = call
	s.mu.Unlock()

	refreshed, err := s.refreshCredential(ctx, cred)

	s.mu.Lock()
	delete(s.refreshes, key)
	call.cred = refreshed
	call.err = err
	close(call.done)
	s.mu.Unlock()

	return refreshed.Clone(), err
}

func (s *Service) refreshCredential(ctx context.Context, cred *Credential) (*Credential, error) {
	var (
		refreshed *Credential
		err       error
	)
	switch normalizeProvider(cred.Provider) {
	case config.OAuthProviderCodex:
		refreshed, err = s.codex.Refresh(ctx, cred)
	default:
		return nil, fmt.Errorf("unsupported oauth provider %q", cred.Provider)
	}
	if err != nil {
		return nil, err
	}
	if err := s.store.Save(refreshed); err != nil {
		return nil, err
	}
	return refreshed, nil
}

func (s *Service) finishSessionWithError(sessionID string, msg string, callback *callbackServer) *LoginSession {
	_ = callback.Close()
	s.mu.Lock()
	defer s.mu.Unlock()
	session := s.sessions[sessionID]
	session.callback = nil
	session.Status = LoginStatusError
	session.Error = strings.TrimSpace(msg)
	return session.Clone()
}

func (s *Service) sweepExpiredSessionsLocked() {
	now := s.now()
	for id, session := range s.sessions {
		if session == nil || session.callback == nil {
			continue
		}
		if !session.ExpiresAt.IsZero() && !session.ExpiresAt.After(now) {
			_ = session.callback.Close()
			session.callback = nil
			session.Status = LoginStatusExpired
			session.Error = "oauth session expired"
			s.sessions[id] = session
		}
	}
}

type callbackResult struct {
	Code  string
	State string
	Error string
}

type callbackServer struct {
	listener net.Listener
	server   *http.Server
	results  chan callbackResult
}

func startCallbackServer(host string, port int, path string) (*callbackServer, string, error) {
	if strings.TrimSpace(host) == "" {
		host = "127.0.0.1"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + strings.TrimSpace(path)
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, "", err
	}

	server := &callbackServer{
		listener: listener,
		results:  make(chan callbackResult, 1),
	}
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		result := callbackResult{
			Code:  strings.TrimSpace(r.URL.Query().Get("code")),
			State: strings.TrimSpace(r.URL.Query().Get("state")),
			Error: strings.TrimSpace(r.URL.Query().Get("error")),
		}
		if result.Error == "" && result.Code == "" {
			result.Error = "authorization code not found"
		}
		select {
		case server.results <- result:
		default:
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><body><script>(function(){if(window.opener){window.close();}})();</script><h1>Authentication received</h1><p>Return to Clipal to finish setup. You can close this window if it does not close automatically.</p></body></html>`))
	})
	server.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		_ = server.server.Serve(listener)
	}()

	addr := listener.Addr().(*net.TCPAddr)
	redirectHost := host
	if redirectHost == "" || redirectHost == "0.0.0.0" || redirectHost == "::" {
		redirectHost = "127.0.0.1"
	}
	redirectURI := "http://" + net.JoinHostPort(redirectHost, strconv.Itoa(addr.Port)) + path
	return server, redirectURI, nil
}

func (s *callbackServer) Poll() (callbackResult, bool) {
	if s == nil {
		return callbackResult{}, false
	}
	select {
	case result := <-s.results:
		return result, true
	default:
		return callbackResult{}, false
	}
}

func (s *callbackServer) Close() error {
	if s == nil || s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if s.listener != nil {
		_ = s.listener.Close()
	}
	return err
}

func randomID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}
