package web

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/lansespirit/Clipal/internal/config"
	oauthpkg "github.com/lansespirit/Clipal/internal/oauth"
)

const (
	maxOAuthImportFiles           = 512
	maxOAuthImportFileBytes       = 1 << 20 // 1 MiB per uploaded credential file
	oauthImportMultipartMaxMemory = 8 << 20 // spill larger payloads to disk
)

type oauthImportCandidate struct {
	cred   *oauthpkg.Credential
	result OAuthImportFileResultResponse
}

func (a *API) HandleImportCLIProxyAPICredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(oauthImportMultipartMaxMemory); err != nil {
		writeError(w, fmt.Sprintf("invalid multipart form: %v", err), http.StatusBadRequest)
		return
	}

	clientType, ok := config.CanonicalClientType(strings.TrimSpace(r.FormValue("client_type")))
	if !ok {
		writeError(w, "invalid client type", http.StatusBadRequest)
		return
	}
	requestedProvider := config.OAuthProvider(strings.ToLower(strings.TrimSpace(r.FormValue("provider"))))
	if requestedProvider == "" {
		writeError(w, "provider is required", http.StatusBadRequest)
		return
	}
	if err := validateOAuthProviderForClient(clientType, requestedProvider); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	headers := r.MultipartForm.File["files"]
	if len(headers) == 0 {
		writeError(w, "no credential files uploaded", http.StatusBadRequest)
		return
	}
	if len(headers) > maxOAuthImportFiles {
		writeError(w, fmt.Sprintf("too many credential files: max %d", maxOAuthImportFiles), http.StatusBadRequest)
		return
	}

	resp := OAuthImportResponse{
		ClientType: clientType,
		Provider:   requestedProvider,
		Results:    make([]OAuthImportFileResultResponse, 0, len(headers)),
	}
	candidates := make([]oauthImportCandidate, 0, len(headers))
	for _, header := range headers {
		candidate := a.parseCLIProxyAPIImportCandidate(header, requestedProvider)
		resp.addResult(candidate.result)
		candidates = append(candidates, candidate)
	}

	a.configMu.Lock()
	defer a.configMu.Unlock()

	cfg := a.loadConfigOrWriteError(w)
	if cfg == nil {
		return
	}
	cc, err := getClientConfigRef(cfg, clientType)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	seen := make(map[string]struct{}, len(candidates))
	changed := false
	for i := range candidates {
		candidate := &candidates[i]
		if candidate.cred == nil {
			continue
		}

		key := string(candidate.cred.Provider) + ":" + candidate.cred.Ref
		if _, exists := seen[key]; exists {
			candidate.result.Status = "skipped"
			candidate.result.Message = "duplicate account in selected files"
			candidate.cred = nil
			resp.recountResult(i, candidate.result)
			continue
		}
		seen[key] = struct{}{}

		if err := a.oauth.Store().Save(candidate.cred); err != nil {
			candidate.result.Status = "failed"
			candidate.result.Message = fmt.Sprintf("save imported credential: %v", err)
			candidate.cred = nil
			resp.recountResult(i, candidate.result)
			continue
		}

		provider, linked := ensureOAuthProviderLinked(cc, candidate.cred)
		candidate.result.Status = "imported"
		candidate.result.ProviderName = provider.Name
		if linked {
			candidate.result.Message = fmt.Sprintf("imported account and created provider %s", provider.Name)
			resp.LinkedCount++
			changed = true
		} else {
			candidate.result.Message = fmt.Sprintf("imported account and reused provider %s", provider.Name)
		}
		resp.recountResult(i, candidate.result)
	}

	if changed {
		if !a.saveClientConfigOrWriteError(w, clientType, cfg) {
			return
		}
	}

	resp.Message = summarizeOAuthImport(resp)
	writeJSON(w, resp)
}

func (a *API) parseCLIProxyAPIImportCandidate(header *multipart.FileHeader, requestedProvider config.OAuthProvider) oauthImportCandidate {
	result := OAuthImportFileResultResponse{
		File:   strings.TrimSpace(header.Filename),
		Status: "skipped",
	}
	if result.File == "" {
		result.File = "credential.json"
	}
	if ext := strings.ToLower(filepath.Ext(result.File)); ext != ".json" {
		result.Message = "skipped non-JSON file"
		return oauthImportCandidate{result: result}
	}

	data, err := readOAuthImportFile(header)
	if err != nil {
		result.Status = "failed"
		result.Message = err.Error()
		return oauthImportCandidate{result: result}
	}
	cred, err := oauthpkg.ParseCLIProxyAPICredential(data)
	if err != nil {
		switch {
		case errors.Is(err, oauthpkg.ErrCLIProxyAPINotCredential):
			result.Message = "skipped file without supported OAuth credential data"
		case errors.Is(err, oauthpkg.ErrCLIProxyAPIUnsupportedType):
			result.Message = err.Error()
		case errors.Is(err, oauthpkg.ErrCLIProxyAPIDisabledCredential):
			result.Message = "skipped disabled OAuth credential"
		default:
			result.Status = "failed"
			result.Message = err.Error()
		}
		return oauthImportCandidate{result: result}
	}
	if cred.Provider != requestedProvider {
		result.Message = fmt.Sprintf("skipped %s credential while importing %s accounts", cred.Provider, requestedProvider)
		return oauthImportCandidate{result: result}
	}

	result.Provider = cred.Provider
	result.Ref = cred.Ref
	result.Email = cred.Email
	return oauthImportCandidate{cred: cred, result: result}
}

func readOAuthImportFile(header *multipart.FileHeader) ([]byte, error) {
	f, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("open uploaded file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	data, err := io.ReadAll(io.LimitReader(f, maxOAuthImportFileBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read uploaded file: %w", err)
	}
	if len(data) > maxOAuthImportFileBytes {
		return nil, fmt.Errorf("uploaded file exceeds %d bytes", maxOAuthImportFileBytes)
	}
	return data, nil
}

func (resp *OAuthImportResponse) addResult(result OAuthImportFileResultResponse) {
	resp.Results = append(resp.Results, result)
	switch result.Status {
	case "imported":
		resp.ImportedCount++
	case "failed":
		resp.FailedCount++
	default:
		resp.SkippedCount++
	}
}

func (resp *OAuthImportResponse) recountResult(index int, next OAuthImportFileResultResponse) {
	if resp == nil || index < 0 || index >= len(resp.Results) {
		return
	}
	prev := resp.Results[index]
	resp.adjustResultCount(prev.Status, -1)
	resp.adjustResultCount(next.Status, 1)
	resp.Results[index] = next
}

func (resp *OAuthImportResponse) adjustResultCount(status string, delta int) {
	switch status {
	case "imported":
		resp.ImportedCount += delta
	case "failed":
		resp.FailedCount += delta
	default:
		resp.SkippedCount += delta
	}
}

func summarizeOAuthImport(resp OAuthImportResponse) string {
	parts := []string{
		fmt.Sprintf("imported %d account(s)", resp.ImportedCount),
		fmt.Sprintf("created %d provider(s)", resp.LinkedCount),
	}
	if resp.SkippedCount > 0 {
		parts = append(parts, fmt.Sprintf("skipped %d file(s)", resp.SkippedCount))
	}
	if resp.FailedCount > 0 {
		parts = append(parts, fmt.Sprintf("failed %d file(s)", resp.FailedCount))
	}
	return strings.Join(parts, ", ")
}
