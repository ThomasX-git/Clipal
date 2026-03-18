package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lansespirit/Clipal/internal/config"
	"github.com/lansespirit/Clipal/internal/logger"
	"github.com/lansespirit/Clipal/internal/notify"
)

func nextProviderName(cp *ClientProxy, fromIndex int) (idx int, name string) {
	idx = cp.nextActiveIndex(fromIndex)
	if idx == fromIndex {
		return idx, ""
	}
	if idx < 0 || idx >= len(cp.providers) {
		return idx, ""
	}
	n := strings.TrimSpace(cp.providers[idx].Name)
	if n == "" {
		return idx, ""
	}
	return idx, n
}

func (cp *ClientProxy) announceProviderSwitch(from string, to string, reason string, status int) {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if from == "" || to == "" || from == to {
		return
	}
	switchView := DescribeProviderSwitch(from, to, reason, status)
	logger.Info("[%s] %s. %s", cp.clientType, switchView.Label, switchView.Detail)
	cp.recordProviderSwitch(from, to, reason, status)
	notify.ProviderSwitched(string(cp.clientType), switchView.Label, switchView.Detail)
}

func unavailableRequestStatus(reason string) (result string, status int, detail string) {
	if reason == "rate_limit" || reason == "overloaded" {
		return "all_providers_unavailable", http.StatusTooManyRequests, "All providers are rate limited; retry later."
	}
	return "all_providers_unavailable", http.StatusServiceUnavailable, "All providers are temporarily unavailable; retry later."
}

func describeRequestBuildFailure(provider string, err error) string {
	msg := "local request setup failed"
	if err != nil {
		msg = truncateString(sanitizeLogString(err.Error()), 512)
	}
	if strings.TrimSpace(provider) == "" {
		return "Request could not be prepared locally: " + msg
	}
	return fmt.Sprintf("%s request could not be prepared locally: %s", provider, msg)
}

// forwardWithFailover forwards the request with automatic failover.
func (cp *ClientProxy) forwardWithFailover(w http.ResponseWriter, req *http.Request, path string) {
	if cp.mode == config.ClientModeManual {
		cp.forwardManual(w, req, path)
		return
	}

	cp.reactivateExpired()

	// If the client has already gone away (or the server is shutting down), don't do any work.
	if err := req.Context().Err(); err != nil {
		return
	}

	// Read the request body once for potential retries
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Error("[%s] failed to read request body: %v", cp.clientType, err)
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			cp.recordTerminalRequest(time.Now(), "", http.StatusRequestEntityTooLarge, "request_rejected", "Request body too large.")
			writeProxyError(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		cp.recordTerminalRequest(time.Now(), "", http.StatusBadRequest, "request_rejected", "Failed to read request body.")
		writeProxyError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	// Atomically get active count and start index to avoid TOCTOU race.
	active, startIndex := cp.getActiveCountAndStartIndex()
	if active == 0 {
		if wait, reason, ok := cp.timeUntilNextAvailable(); ok && wait > 0 {
			result, status, detail := unavailableRequestStatus(reason)
			cp.recordTerminalRequest(time.Now(), "", status, result, detail)
		} else {
			cp.recordTerminalRequest(time.Now(), "", http.StatusServiceUnavailable, "all_providers_unavailable", "All providers are unavailable.")
		}
		if handled := cp.handleAllUnavailable(w); handled {
			return
		}
		logger.Error("[%s] all providers unavailable", cp.clientType)
		writeProxyError(w, "All providers are unavailable", http.StatusServiceUnavailable)
		return
	}

	attempted := 0
	lastSwitchReason := ""
	lastSwitchStatus := 0
	lastFailedProvider := ""
	attemptSummaries := make([]string, 0, active)
	hadUpstreamAttempt := false

	for offset := 0; offset < len(cp.providers) && attempted < active; offset++ {
		if err := req.Context().Err(); err != nil {
			return
		}

		index := (startIndex + offset) % len(cp.providers)
		if cp.isDeactivated(index) {
			continue
		}
		now := time.Now()
		allow := cp.allowCircuit(now, index)
		if !allow.allowed {
			continue
		}
		provider := cp.providers[index]

		attempted++
		logger.Debug("[%s] forwarding to: %s (attempt %d/%d)", cp.clientType, provider.Name, attempted, active)

		attemptCtx, cancelAttempt := context.WithCancelCause(req.Context())

		// Create the proxy request
		reqWithAttemptCtx := req.WithContext(attemptCtx)
		proxyReq, err := cp.createProxyRequest(reqWithAttemptCtx, provider, path, bodyBytes)
		if err != nil {
			summary := describeRequestBuildFailure(provider.Name, err)
			attemptSummaries = append(attemptSummaries, summary)
			lastFailedProvider = provider.Name
			logger.Error("[%s] %s", cp.clientType, summary)
			cp.releaseCircuitPermit(index, allow.usedProbe)
			cancelAttempt(nil)
			continue
		}
		hadUpstreamAttempt = true

		// Ensure the body can be retried (http.Request may be reused by the transport).
		proxyReq.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}

		// Send the request
		resp, err := cp.httpClient.Do(proxyReq)
		if err != nil {
			// Don't retry across providers when the request context is already canceled;
			// this otherwise produces misleading "all providers failed" logs.
			if req.Context().Err() != nil {
				cp.releaseCircuitPermit(index, allow.usedProbe)
				cancelAttempt(nil)
				return
			}
			cp.recordCircuitFailure(time.Now(), index, allow.usedProbe, "network")
			cancelAttempt(nil)
			nextIndex, nextName := nextProviderName(cp, index)
			summary := describeAttemptFailure(provider.Name, "network", 0, true)
			attemptSummaries = append(attemptSummaries, summary)
			if nextName != "" {
				logger.Warn("[%s] %s; trying next=%s", cp.clientType, summary, nextName)
			} else {
				logger.Warn("[%s] %s; trying next provider", cp.clientType, summary)
			}
			lastSwitchReason = "network"
			lastSwitchStatus = 0
			lastFailedProvider = provider.Name
			if nextName != "" {
				cp.announceProviderSwitch(provider.Name, nextName, lastSwitchReason, lastSwitchStatus)
			}
			cp.setCurrentIndex(nextIndex)
			continue
		}

		var (
			action   failureAction
			reason   string
			msg      string
			cooldown time.Duration
		)
		inspect := resp.StatusCode == http.StatusUnauthorized ||
			resp.StatusCode == http.StatusForbidden ||
			resp.StatusCode == http.StatusPaymentRequired ||
			resp.StatusCode == http.StatusTooManyRequests ||
			shouldRetry(resp.StatusCode)
		if inspect {
			body, truncated := readResponseBodyBytes(resp, 32*1024)
			action, reason, msg, cooldown = classifyUpstreamFailure(resp.StatusCode, resp.Header, body, truncated)
		} else {
			action = failureReturnToClient
		}

		if action != failureReturnToClient {
			cp.recordCircuitFailureFromClassification(time.Now(), index, allow.usedProbe, reason)
			resp.Body.Close()
			cancelAttempt(nil)
			lastSwitchReason = reason
			lastSwitchStatus = resp.StatusCode
			lastFailedProvider = provider.Name
			nextIndex, nextName := nextProviderName(cp, index)
			summary := describeAttemptFailure(provider.Name, reason, resp.StatusCode, false)
			attemptSummaries = append(attemptSummaries, summary)
			switch action {
			case failureDeactivateAndRetryNext:
				cp.deactivateFor(index, reason, resp.StatusCode, msg, cp.reactivateAfter)
				if nextName != "" {
					logger.Error("[%s] %s; marking provider unavailable and trying next=%s", cp.clientType, summary, nextName)
				} else {
					logger.Error("[%s] %s; marking provider unavailable and trying next provider", cp.clientType, summary)
				}
			case failureRetryNext:
				summaryWithCooldown := summary
				if cooldown > 0 {
					cp.deactivateFor(index, reason, resp.StatusCode, msg, cooldown)
					summaryWithCooldown = fmt.Sprintf("%s; cooling down for %s", summary, cooldown)
				}
				if nextName != "" {
					logger.Warn("[%s] %s; trying next=%s", cp.clientType, summaryWithCooldown, nextName)
				} else {
					logger.Warn("[%s] %s; trying next provider", cp.clientType, summaryWithCooldown)
				}
			}
			if nextName != "" {
				cp.announceProviderSwitch(provider.Name, nextName, lastSwitchReason, lastSwitchStatus)
			}
			cp.setCurrentIndex(nextIndex)
			continue
		}

		// Success (or pass-through response) - copy response to client. For streaming endpoints
		// (SSE), wait for the first body bytes before sending headers so we can fail over cleanly
		// if the upstream hangs after headers.
		// Found a working provider.
		onCommit := func() {
			cp.setCurrentIndex(index)
		}

		result := cp.streamResponseToClient(w, resp, req, attemptCtx, cancelAttempt, index, allow, onCommit)
		if result.kind == streamFinal {
			cp.logRequestResult(provider.Name, resp.StatusCode, result, false)
			return
		}

		// Failed before committing headers (e.g. idle timeout during first byte read).
		// Failover to next provider.
		if isUpstreamIdleTimeout(attemptCtx, attemptCtx.Err()) {
			summary := describeAttemptFailure(provider.Name, "idle_timeout", 0, true)
			attemptSummaries = append(attemptSummaries, summary)
			logger.Warn("[%s] %s; trying next provider", cp.clientType, summary)
			lastSwitchReason = "idle_timeout"
			lastSwitchStatus = 0
			lastFailedProvider = provider.Name
			cp.recordCircuitFailure(time.Now(), index, allow.usedProbe, "idle_timeout")
		} else {
			summary := describeAttemptFailure(provider.Name, "network", 0, true)
			attemptSummaries = append(attemptSummaries, summary)
			logger.Warn("[%s] %s; trying next provider", cp.clientType, summary)
			lastSwitchReason = "network"
			lastSwitchStatus = 0
			lastFailedProvider = provider.Name
			cp.recordCircuitFailure(time.Now(), index, allow.usedProbe, "network")
		}
		cancelAttempt(nil)
		nextIndex, nextName := nextProviderName(cp, index)
		if nextName != "" {
			cp.announceProviderSwitch(provider.Name, nextName, lastSwitchReason, lastSwitchStatus)
		}
		cp.setCurrentIndex(nextIndex)
		continue
	}

	// If we've cooled down all providers during this request, surface a Retry-After to the client.
	if cp.activeProviderCount() == 0 {
		if wait, reason, ok := cp.timeUntilNextAvailable(); ok && wait > 0 {
			result, status, detail := unavailableRequestStatus(reason)
			cp.recordTerminalRequest(time.Now(), "", status, result, detail)
		}
		if handled := cp.handleAllUnavailable(w); handled {
			return
		}
	}

	lastProvider := strings.TrimSpace(lastFailedProvider)
	terminalResult := "all_providers_failed"
	terminalStatus := http.StatusServiceUnavailable
	if !hadUpstreamAttempt {
		terminalResult = "request_rejected"
		terminalStatus = http.StatusBadGateway
	}
	cp.recordTerminalRequest(time.Now(), lastProvider, terminalStatus, terminalResult, strings.Join(attemptSummaries, "; "))
	if len(attemptSummaries) > 0 {
		logger.Error("[%s] all providers failed: %s", cp.clientType, strings.Join(attemptSummaries, "; "))
	} else {
		logger.Error("[%s] all providers failed", cp.clientType)
	}
	writeProxyError(w, "All providers failed", http.StatusServiceUnavailable)
}

// forwardCountTokensWithFailover forwards Claude Code /v1/messages/count_tokens requests while
// keeping the main conversation provider sticky (cp.currentIndex) unchanged.
//
// Rationale: Claude Code calls count_tokens frequently; using those failures to move the primary
// provider can reduce context-cache effectiveness and increase token usage.
func (cp *ClientProxy) forwardCountTokensWithFailover(w http.ResponseWriter, req *http.Request, path string) {
	if cp.mode == config.ClientModeManual {
		cp.forwardManual(w, req, path)
		return
	}

	cp.reactivateExpired()

	if err := req.Context().Err(); err != nil {
		return
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Error("[%s] failed to read request body: %v", cp.clientType, err)
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeProxyError(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		writeProxyError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	// Atomically get active count and start index to avoid TOCTOU race.
	active, startIndex := cp.getActiveCountAndCountTokensStartIndex()
	if active == 0 {
		if handled := cp.handleAllUnavailable(w); handled {
			return
		}
		logger.Error("[%s] all providers unavailable", cp.clientType)
		writeProxyError(w, "All providers are unavailable", http.StatusServiceUnavailable)
		return
	}

	attempted := 0
	attemptSummaries := make([]string, 0, active)
	hadUpstreamAttempt := false

	for offset := 0; offset < len(cp.providers) && attempted < active; offset++ {
		if err := req.Context().Err(); err != nil {
			return
		}

		index := (startIndex + offset) % len(cp.providers)
		if cp.isDeactivated(index) {
			continue
		}
		now := time.Now()
		allow := cp.allowCircuit(now, index)
		if !allow.allowed {
			continue
		}
		attempted++
		provider := cp.providers[index]

		logger.Debug("[%s] forwarding to: %s (count_tokens attempt %d/%d)", cp.clientType, provider.Name, attempted, active)

		attemptCtx, cancelAttempt := context.WithCancelCause(req.Context())
		reqWithAttemptCtx := req.WithContext(attemptCtx)

		proxyReq, err := cp.createProxyRequest(reqWithAttemptCtx, provider, path, bodyBytes)
		if err != nil {
			summary := describeRequestBuildFailure(provider.Name, err)
			attemptSummaries = append(attemptSummaries, summary)
			logger.Error("[%s] %s during count_tokens", cp.clientType, summary)
			cp.setCountTokensIndex(cp.nextActiveIndex(index))
			cp.releaseCircuitPermit(index, allow.usedProbe)
			cancelAttempt(nil)
			continue
		}
		hadUpstreamAttempt = true

		proxyReq.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}

		resp, err := cp.httpClient.Do(proxyReq)
		if err != nil {
			if req.Context().Err() != nil {
				cp.releaseCircuitPermit(index, allow.usedProbe)
				cancelAttempt(nil)
				return
			}
			cp.recordCircuitFailure(time.Now(), index, allow.usedProbe, "network")
			cancelAttempt(nil)
			nextIndex, nextName := nextProviderName(cp, index)
			summary := describeAttemptFailure(provider.Name, "network", 0, true)
			attemptSummaries = append(attemptSummaries, summary)
			if nextName != "" {
				logger.Warn("[%s] %s during count_tokens; trying next=%s", cp.clientType, summary, nextName)
			} else {
				logger.Warn("[%s] %s during count_tokens; trying next provider", cp.clientType, summary)
			}
			cp.setCountTokensIndex(nextIndex)
			continue
		}

		// For count_tokens, only treat auth/billing failures as hard signals that can deactivate a provider.
		// Other transient failures should not impact the main conversation stickiness (cp.currentIndex).
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			msg := readAndTruncateResponse(resp, 2048)
			resp.Body.Close()
			cp.deactivateFor(index, "auth", resp.StatusCode, msg, cp.reactivateAfter)
			cp.recordCircuitFailureFromClassification(time.Now(), index, allow.usedProbe, "auth")
			nextIndex, nextName := nextProviderName(cp, index)
			cp.setCountTokensIndex(nextIndex)
			summary := describeAttemptFailure(provider.Name, "auth", resp.StatusCode, false)
			attemptSummaries = append(attemptSummaries, summary)
			if nextName != "" {
				logger.Error("[%s] %s during count_tokens; marking provider unavailable and trying next=%s", cp.clientType, summary, nextName)
			} else {
				logger.Error("[%s] %s during count_tokens; marking provider unavailable and trying next provider", cp.clientType, summary)
			}
			cancelAttempt(nil)
			continue
		}
		if resp.StatusCode == http.StatusPaymentRequired {
			msg := readAndTruncateResponse(resp, 2048)
			resp.Body.Close()
			cp.deactivateFor(index, "billing", resp.StatusCode, msg, cp.reactivateAfter)
			cp.recordCircuitFailureFromClassification(time.Now(), index, allow.usedProbe, "billing")
			nextIndex, nextName := nextProviderName(cp, index)
			cp.setCountTokensIndex(nextIndex)
			summary := describeAttemptFailure(provider.Name, "billing", resp.StatusCode, false)
			attemptSummaries = append(attemptSummaries, summary)
			if nextName != "" {
				logger.Error("[%s] %s during count_tokens; marking provider unavailable and trying next=%s", cp.clientType, summary, nextName)
			} else {
				logger.Error("[%s] %s during count_tokens; marking provider unavailable and trying next provider", cp.clientType, summary)
			}
			cancelAttempt(nil)
			continue
		}
		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			_ = readAndTruncateResponse(resp, 2048)
			resp.Body.Close()
			if resp.StatusCode >= 500 {
				cp.recordCircuitFailureFromClassification(time.Now(), index, allow.usedProbe, "server")
			} else {
				cp.recordCircuitFailureFromClassification(time.Now(), index, allow.usedProbe, "rate_limit")
			}
			nextIndex, nextName := nextProviderName(cp, index)
			cp.setCountTokensIndex(nextIndex)
			reason := "rate_limit"
			if resp.StatusCode >= 500 {
				reason = "server"
			}
			summary := describeAttemptFailure(provider.Name, reason, resp.StatusCode, false)
			attemptSummaries = append(attemptSummaries, summary)
			if nextName != "" {
				logger.Warn("[%s] %s during count_tokens; trying next=%s", cp.clientType, summary, nextName)
			} else {
				logger.Warn("[%s] %s during count_tokens; trying next provider", cp.clientType, summary)
			}
			cancelAttempt(nil)
			continue
		}

		// Success (or any non-retriable response) - return to client and make count_tokens sticky.
		onCommit := func() {
			cp.setCountTokensIndex(index)
		}

		result := cp.streamResponseToClient(w, resp, req, attemptCtx, cancelAttempt, index, allow, onCommit)
		if result.kind == streamFinal {
			return
		}

		// Failed before committing headers (count_tokens: try next provider).
		cancelAttempt(nil)
		if req.Context().Err() != nil {
			cp.releaseCircuitPermit(index, allow.usedProbe)
			return
		}
		reason := "network"
		if isUpstreamIdleTimeout(attemptCtx, attemptCtx.Err()) {
			reason = "idle_timeout"
			cp.recordCircuitFailure(time.Now(), index, allow.usedProbe, "idle_timeout")
		} else {
			cp.recordCircuitFailure(time.Now(), index, allow.usedProbe, "network")
		}
		summary := describeAttemptFailure(provider.Name, reason, 0, true)
		attemptSummaries = append(attemptSummaries, summary)
		logger.Warn("[%s] %s during count_tokens; trying next provider", cp.clientType, summary)
		cp.setCountTokensIndex(cp.nextActiveIndex(index))
		continue
	}

	if len(attemptSummaries) > 0 {
		logger.Error("[%s] all providers failed during count_tokens: %s", cp.clientType, strings.Join(attemptSummaries, "; "))
	} else if !hadUpstreamAttempt {
		logger.Error("[%s] count_tokens request could not be prepared for any provider", cp.clientType)
	} else {
		logger.Error("[%s] all providers failed (count_tokens)", cp.clientType)
	}
	writeProxyError(w, "All providers failed", http.StatusServiceUnavailable)
}
