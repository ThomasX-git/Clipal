package proxy

import (
	"encoding/json"
	"mime"
	"net/http"
	"strings"

	"github.com/lansespirit/Clipal/internal/config"
)

func applyProviderRequestOverrides(original *http.Request, requestCtx RequestContext, provider config.Provider, body []byte) []byte {
	if len(body) == 0 || !hasProviderRequestOverrides(provider) || !isJSONRequest(original) {
		return body
	}

	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil || root == nil {
		return body
	}

	if !applyProviderRequestOverridesToRoot(root, requestCtx, provider) {
		return body
	}

	rewritten, err := json.Marshal(root)
	if err != nil {
		return body
	}
	return rewritten
}

func hasProviderRequestOverrides(provider config.Provider) bool {
	return strings.TrimSpace(provider.Model) != "" ||
		strings.TrimSpace(provider.ReasoningEffort) != "" ||
		provider.ThinkingBudgetTokens > 0
}

func isJSONRequest(req *http.Request) bool {
	if req == nil {
		return false
	}
	mediaType := strings.TrimSpace(req.Header.Get("Content-Type"))
	if mediaType == "" {
		return false
	}
	parsed, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		parsed = mediaType
	}
	parsed = strings.ToLower(strings.TrimSpace(parsed))
	return parsed == "application/json" || strings.HasSuffix(parsed, "+json")
}

func applyProviderRequestOverridesToRoot(root map[string]any, requestCtx RequestContext, provider config.Provider) bool {
	changed := false
	model := strings.TrimSpace(provider.Model)
	if model != "" {
		switch requestCtx.Family {
		case ProtocolFamilyOpenAI:
			if isOpenAIGenerationCapability(requestCtx.Capability) {
				root["model"] = model
				changed = true
			}
		case ProtocolFamilyClaude:
			if requestCtx.Capability == CapabilityClaudeMessages || requestCtx.Capability == CapabilityClaudeCountTokens {
				root["model"] = model
				changed = true
			}
		}
	}

	reasoningEffort := strings.TrimSpace(provider.ReasoningEffort)
	if reasoningEffort != "" && requestCtx.Family == ProtocolFamilyOpenAI {
		switch requestCtx.Capability {
		case CapabilityOpenAIResponses:
			reasoning, _ := root["reasoning"].(map[string]any)
			if reasoning == nil {
				reasoning = make(map[string]any)
			}
			reasoning["effort"] = reasoningEffort
			root["reasoning"] = reasoning
			changed = true
		default:
			if _, ok := root["reasoning_effort"]; ok {
				root["reasoning_effort"] = reasoningEffort
				changed = true
			}
		}
	}

	if provider.ThinkingBudgetTokens > 0 &&
		(requestCtx.Capability == CapabilityClaudeMessages || requestCtx.Capability == CapabilityClaudeCountTokens) {
		root["thinking"] = map[string]any{
			"type":          "enabled",
			"budget_tokens": provider.ThinkingBudgetTokens,
		}
		changed = true
	}

	return changed
}

func isOpenAIGenerationCapability(capability RequestCapability) bool {
	switch capability {
	case CapabilityOpenAIResponses, CapabilityOpenAIChatCompletions, CapabilityOpenAICompletions:
		return true
	default:
		return false
	}
}
