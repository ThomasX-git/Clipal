# Session-Sticky Busy Overflow Routing

Date: 2026-03-23

## Summary

Upgrade Clipal routing from failure-driven provider switching to session-sticky routing with black-box congestion handling.

The new behavior treats concurrency-limit `429` responses as a soft-capacity signal instead of a provider health failure:

- keep a session bound to one provider when possible
- pause and retry briefly when that provider reports concurrency saturation
- avoid sending a stampede of new requests to the same saturated provider
- overflow the affected session to the next candidate provider only after controlled retry
- rebind the session to the overflow target so later requests stay sticky there

This change must not degrade existing auth/quota/circuit-breaker semantics.

## Goals

- Preserve same-session affinity to maximize upstream cache reuse.
- Stop interpreting concurrency saturation as provider breakage.
- Support black-box upstreams where actual concurrency limits are unknown and may differ per provider.
- Reduce stampedes when one provider begins returning concurrency-limit `429`.
- Keep current provider failover behavior for hard failures and general transient errors.

## Non-Goals

- Do not infer a fake universal session identifier when the protocol does not provide one.
- Do not implement cluster-wide shared affinity state across multiple Clipal processes.
- Do not add queue persistence across process restarts.
- Do not replace the existing circuit breaker with a generic load balancer.

## Current Problem

Today a `429` classified as `rate_limit` causes provider-level cooldown in the normal failover path. That behavior is reasonable for quota or long reset windows, but it is wrong for concurrency saturation:

- the provider is still healthy
- only some requests should wait or spill
- current logic moves the global routing cursor away from the provider
- later requests lose affinity and may scatter across providers

The runtime has health-oriented state only:

- provider deactivation
- key deactivation
- circuit breaker
- scope-local current provider cursors

It does not model:

- session affinity ownership
- provider busy/backpressure state
- controlled retry before spillover

## External API Reality

There is no universal cross-provider session identifier. Sticky extraction must be capability-specific and layered.

Reusable explicit linkage keys:

- OpenAI `Responses`: request `previous_response_id`, matched against prior response `id`
- Anthropic tool containers: request `container`
- Gemini cached context: request `cached_content`

Reusable cache affinity hints:

- OpenAI `prompt_cache_key`

Common stateless flows still lack a durable session identifier:

- OpenAI `chat/completions`
- Anthropic standard `messages`
- Gemini plain `generateContent`

For those stateless flows, Clipal can still derive a short-lived dynamic conversation feature from the human-message sequence. Response IDs alone should not be treated as a universal session key, but they remain useful as short-lived lookup material for chained APIs such as `previous_response_id`.

## High-Level Design

Add a new routing layer between scope selection and upstream attempt execution:

1. derive an optional sticky key from the incoming request
2. resolve the preferred provider for that sticky key
3. consult provider/key health and provider busy state
4. if preferred provider is busy, wait and retry in a controlled way
5. if retry still encounters concurrency saturation, overflow to the next candidate
6. rebind the sticky key to the overflow target
7. continue normal sticky routing on later requests

This introduces a new state category:

- `busy`: provider has recently reported concurrency saturation and should receive only limited probe traffic until the busy window expires

`busy` is not `deactivated`.

## Runtime State Model

### 1. Sticky Session Bindings

Maintain per-client, per-scope sticky bindings:

- key: `routingScope + stickyKey`
- value:
  - provider index
  - key index if selection is key-specific
  - bound at
  - last seen at
  - source type
  - overflow generation count

Bindings are in-memory and expire after idle TTL.

Recommended defaults:

- explicit linkage key idle TTL: `30m`
- cache affinity key idle TTL: `10m`
- dynamic conversation feature idle TTL: `10m`

Only explicit linkage keys are allowed to trigger durable rebinding. Cache affinity and dynamic conversation features can influence preferred provider choice, but must stay bounded by shorter TTL and cache capacity.

### 2. Response Lookup Cache

Maintain a short-lived map from response-like IDs to provider ownership so chained APIs can recover affinity.

- key: provider-specific response object ID
- value:
  - provider index
  - scope
  - observed at

Recommended TTL:

- `15m`

Usage:

- if a new request references a known object such as `previous_response_id`, resolve affinity directly
- if a response returns a reusable object ID, store it for future lookup
- otherwise ignore

### 3. Dynamic Feature Cache

Maintain a bounded LRU-style cache for dynamic conversation features used by stateless request flows.

- key:
  - client type
  - routing scope
  - model name if present
  - normalized dynamic feature
- value:
  - provider index
  - key index if useful
  - last seen at
  - source label

Recommended defaults:

- max entries per client+scope: `1024`
- idle TTL: `10m`
- eviction: least recently used

This cache is intentionally lossy. Newer features push out older ones automatically.

### 4. Busy State

Maintain provider-level busy state, separate from cooldown and circuit breaker:

- `busyUntil`
- `busyBackoffStep`
- `lastBusyReason`
- `probeInFlight`
- optional per-key busy state if a provider has multiple keys

Recommended default behavior:

- first concurrency-limit `429` sets busy window to `5s`
- second consecutive busy signal extends to `10s`
- cap busy backoff at `10s`
- allow only `1` probe request per provider while leaving busy state

Busy state decays on success:

- a successful response from the provider clears busy window and resets backoff step

Concurrency update rule:

- protect busy state with the existing proxy mutex
- updates must only extend the busy window, never shorten it
- when multiple concurrent requests report busy at once, merge by taking:
  - the larger backoff step
  - the later `busyUntil`
- requests that lose the race to update busy state should reuse the already-published `busyUntil` instead of recalculating a shorter one

## Sticky Key Extraction

Implement capability-specific extractors. Extraction returns:

- key string
- level: `L1`, `L2`, or `L3`
- source label for observability

### L1: Explicit Linkage Keys

- OpenAI `Responses`
  - request `previous_response_id`, resolved by matching against prior response `id`
- Anthropic messages with tool container reuse
  - request `container`
- Gemini cached context
  - request `cached_content`

`L1` keys are strong enough for durable session rebinding.

### L2: Explicit Cache Affinity Keys

- OpenAI `Responses`
  - request `prompt_cache_key`

`L2` keys are not true conversation IDs, but they are stable enough to keep cache-sensitive flows on one provider when possible.

### L3: Dynamic Conversation Feature Keys

For stateless request flows with message history but without an explicit reusable key, derive a short-lived feature from the human-message sequence.

Rules:

- only inspect human-authored turns
- ignore assistant/model reply turns when building the feature
- for response-side feature recording:
  - use the last human message in the completed conversation
- for request-side feature lookup:
  - use the second-to-last human message
  - reason: the last human message is usually the new user turn being answered now, while the second-to-last human message anchors the existing conversation prefix
- if there is only one human message in the request, treat it as a first-turn request and do not create `L3` affinity
- normalize extracted text before keying:
  - trim whitespace
  - collapse internal whitespace
  - lowercase
- compute a hash from the full normalized human text
- keep the first `24` characters of the normalized human text only as an observability preview
- include capability and model name when present to reduce accidental collisions

Examples:

- request with human turns `[H1, H2, H3]`
  - request-side feature uses `H2`
  - after the response completes, response-side feature recorded for future matching uses `H3`
- request with human turns `[H1]`
  - no `L3` feature recorded or matched

`L3` keys are heuristic and must remain short-lived and capacity-bounded.

### No Extractor

For capabilities without an explicit key or usable human-message structure, return no sticky key.

Fallback behavior in that case:

- use existing scope-local current-provider behavior
- apply busy-aware spillover to the current request
- do not create long-lived session ownership

## Concurrency-Limit Classification

Split current `429` handling into:

- hard unavailable
  - auth masquerading as `429`
  - quota exhaustion
  - long reset windows that should still deactivate key/provider
- soft busy
  - concurrency saturation
  - short-lived overload where retry is expected to work soon

Add a new classification result:

- `failureBusyRetry`

Classification rules:

1. preserve current auth/quota detection first
2. treat explicit concurrency phrases as a prerequisite for `busy`
3. treat only very short `Retry-After` windows as `busy`
4. keep existing `rate_limit` cooldown behavior for short windows that do not also carry concurrency semantics
5. keep existing `rate_limit` cooldown behavior for long reset windows

Examples of `busy` indicators:

- error message contains `concurrent`
- error message contains `too many concurrent`
- error message contains `capacity`
- provider-specific overload types when semantics imply near-term retry
- `Retry-After <= 3s`

Busy classification should require both:

- a concurrency/near-term-capacity signal in the body or provider-specific error type
- a short retry hint if a retry hint is present

If concurrency wording is absent, Clipal must fall back to the existing cooldown logic even when `Retry-After` is short.

Examples of normal cooldown indicators:

- `Retry-After` much larger than the busy cap
- short `Retry-After` without concurrency wording
- explicit rate/token/request budget exhaustion without concurrency wording

This split lets Clipal avoid poisoning provider health when the upstream only wants brief load shedding.

## Request Handling State Machine

### Path A: No Sticky Key

1. use current scope cursor to choose a preferred provider
2. if provider is not busy, attempt normally
3. if provider is busy, wait for the busy window only once for this request
4. retry on the same provider through the busy probe gate
5. if busy again, spill to next candidate for this request only
6. do not create long-lived session binding

### Path B: Strong Sticky Key Exists

1. resolve bound provider from sticky map if present
2. otherwise choose provider using current scope cursor and create binding
3. if bound provider is busy, queue the request behind its busy window
4. when busy window elapses, retry through the provider busy probe gate
5. if request succeeds, keep binding unchanged
6. if request returns busy again after controlled retry, overflow to next candidate
7. on successful overflow, rebind the sticky key to the new provider

### Path C: Cache-Level Key Exists

1. prefer historical provider if known
2. never rely on `L2` or `L3` key as durable ownership equal to `L1`
3. allow overflow-driven rebinding into the bounded affinity caches
4. expire or evict old entries automatically
5. if a later `L1` key appears, it overrides prior `L2`/`L3` affinity

## Controlled Queueing

Queueing should be local and bounded by the request context.

Rules:

- waiting is an inline wait inside the same `forwardWithFailover` request goroutine
- the wait path reuses the already-buffered `bodyBytes`; retry does not reread `req.Body`
- do not send immediate retries while `busyUntil` is in the future
- waiting must respect client cancellation and request deadline
- only one probe request per provider should be allowed to test recovery when the busy window expires
- other requests should continue waiting briefly or spill, depending on remaining context budget
- there is no background queue and no synthetic `429` returned just to ask the client to retry
- maximum proxy-side wait per request should be bounded

Recommended default maximum inline wait:

- `8s`

If the remaining wait budget would exceed the maximum inline wait, Clipal should overflow immediately instead of holding the request longer.

Recommended default request behavior:

- wait until `busyUntil`
- attempt one probe on the preferred provider
- if probe returns busy again, update busy window to next step and overflow immediately

This achieves the desired user experience:

- `429 busy -> wait 5s -> retry -> if still busy, overflow`
- later requests in the same session stick to the overflow target
- retries happen within the same proxied request using buffered request bytes

## Spillover Selection

Overflow selection should reuse the existing scope-aware candidate order, excluding:

- disabled providers
- deactivated providers
- providers with no active keys
- providers blocked by open circuit
- providers currently in an active busy window, unless every remaining candidate is also busy

Selection order:

1. continue using current provider priority ordering
2. prefer non-busy candidates
3. if all remaining candidates are busy, choose the one with the earliest `busyUntil`

This preserves existing admin expectations around provider order.

## Interaction With Multi-Key Providers

Busy handling should be key-aware before escalating to provider-wide busy.

Recommended behavior:

1. if a request on one key returns busy, mark that key busy for the same busy window
2. try another active key in the same provider if available
3. only mark the provider busy when all currently usable keys are busy or the provider-level signal is clearly shared

This preserves Clipal's current "retry next key before next provider" behavior.

## Feature Extraction Details

### Request/Response Pairing

The affinity system must distinguish between:

- request-side lookup features
- response-side learning features

This is necessary because stateless message APIs send the full transcript on every request.

For `L3` dynamic conversation features:

- request-side lookup uses the second-to-last human message
- response-side learning uses the last human message after a successful completion
- this includes first-turn requests with exactly one human message
- that first-turn learning is required so the next request with human turns `[H1, H2]` can look up by `H1`

That pairing lets Clipal connect:

- request `n + 1` back to the provider that answered request `n`

without accidentally keying on the fresh user turn that has not yet produced a cached response.

### Response-Side Cache Writes

Cache writes should happen only after the upstream response is considered successful enough to commit affinity learning.

Recommended write timing:

- run response-side affinity/cache writes from the successful completion path in [internal/proxy/failover_stream.go](/Users/sean/Programs/Clipal/internal/proxy/failover_stream.go)
- invoke the learning hook only after Clipal has decided the response completed successfully
- do not learn from failed, partial, interrupted, or overflow-aborted attempts

Write rules by event:

- response contains a reusable response object `id`
  - write `response id -> provider` into Response Lookup Cache
- response succeeds and the request body exposes a learnable `L3` feature
  - write `dynamic feature hash -> provider` into Dynamic Feature Cache
- request already used `previous_response_id`
  - no additional special `L1` write is required beyond maintaining the existing sticky binding

This keeps the two caches separate:

- Response Lookup Cache supports explicit chaining like `previous_response_id`
- Dynamic Feature Cache supports stateless conversation heuristics

### Supported Message Shapes

Message inspection should support only narrow, explicit shapes at first:

- OpenAI-style `messages`
- Anthropic-style `messages`
- Gemini `contents`

If a payload shape cannot be parsed confidently, return no `L3` key.

### Collision Handling

`L3` keys are heuristic and collisions are possible. Mitigations:

- include client type, scope, and model in the cache key
- use the hash of the full normalized human text as the actual key
- use the first `24` normalized characters only for logs and UI
- use only bounded TTL and LRU capacity
- never elevate `L3` to the same authority as `L1`

## Reload Behavior

Hot reload should preserve routing continuity when provider runtime identity remains compatible.

Recommended inheritance policy:

- sticky `L1` bindings:
  - inherit by provider name when `sameProviderRuntimeIdentity` still matches
- `L2` cache-affinity entries:
  - inherit by provider name when `sameProviderRuntimeIdentity` still matches
- response lookup cache:
  - inherit by provider name when `sameProviderRuntimeIdentity` still matches
- `L3` dynamic feature cache:
  - inherit by provider name when `sameProviderRuntimeIdentity` still matches
- busy state:
  - inherit by provider name when `sameProviderRuntimeIdentity` still matches

If a provider disappears or its runtime identity changes materially, all inherited affinity/busy state for that provider must be dropped instead of remapped.

## Interaction With Existing Health Logic

Keep current semantics unchanged for:

- `401` / `403` auth failures
- `402` billing failures
- quota-style `429`
- transport failures
- `5xx`
- idle timeout
- incomplete protocol stream

Mapping:

- auth/billing/quota -> deactivation as today
- network/server/idle/protocol -> circuit breaker as today
- concurrency busy -> busy state only

Busy events must not:

- advance the global scope cursor permanently
- open the circuit breaker
- mark provider deactivated

The cursor should only move permanently on successful overflow rebinding or on existing hard-failure logic.

## Observability

Extend runtime status and logs with:

- sticky key source labels
- sticky key level
- session binding counts
- dynamic feature cache size
- provider busy state
- key busy counts
- last busy event
- whether last provider switch was `hard_failover` or `busy_overflow`

Suggested status additions:

- provider snapshot:
  - `busy_until`
  - `busy_backoff`
  - `busy_probe_inflight`
- client snapshot:
  - `sticky_binding_count`
  - `response_lookup_count`
  - `dynamic_feature_cache_count`

Suggested log lines:

- `sticky bind scope=openai_responses key_source=previous_response_id provider=p1`
- `sticky learn scope=openai_chat_completions key_level=L3 feature=\"how to deploy to clou\" provider=p1`
- `provider busy wait provider=p1 wait=5s inline=true`
- `provider busy provider=p1 retry_in=5s reason=concurrency_limit`
- `session overflow key_source=previous_response_id from=p1 to=p2 after_retry=1`

## Configuration

Add a small, explicit config block rather than hard-coding behavior:

```yaml
routing:
  sticky_sessions:
    enabled: true
    explicit_ttl: 30m
    cache_hint_ttl: 10m
    dynamic_feature_ttl: 10m
    dynamic_feature_capacity: 1024
    response_lookup_ttl: 15m
  busy_backpressure:
    enabled: true
    retry_delays:
      - 5s
      - 10s
    probe_max_inflight: 1
    short_retry_after_max: 3s
    max_inline_wait: 8s
```

Defaults should preserve current behavior when feature flags are disabled.

## Implementation Outline

### Config

- add routing config structs and defaults in [internal/config/config.go](/Users/sean/Programs/Clipal/internal/config/config.go)
- validate durations and array shape

### Protocol Extraction

- add sticky extraction helpers in [internal/proxy/protocols.go](/Users/sean/Programs/Clipal/internal/proxy/protocols.go) or a new focused file
- parse request body only once and expose a lightweight inspection path for JSON APIs
- add request-side and response-side feature extraction helpers for `L3`
- add explicit response-side learning hooks so successful responses can write:
  - response IDs into Response Lookup Cache
  - learned `L3` hashes into Dynamic Feature Cache

### Runtime State

- extend `ClientProxy` in [internal/proxy/proxy.go](/Users/sean/Programs/Clipal/internal/proxy/proxy.go) with:
  - sticky bindings
  - response lookup cache
  - dynamic feature cache with LRU eviction
  - busy state
  - reload inheritance helpers for sticky/busy/cache state

### Classification

- extend [internal/proxy/failover_classify.go](/Users/sean/Programs/Clipal/internal/proxy/failover_classify.go) with `failureBusyRetry`
- split short concurrency busy from quota/rate-limit cooldown

### Forwarding Logic

- refactor [internal/proxy/failover_forward.go](/Users/sean/Programs/Clipal/internal/proxy/failover_forward.go) to:
  - resolve sticky provider before iteration
  - consult busy state before sending
  - perform bounded inline wait and probe using buffered request bytes
  - overflow and rebind on repeated busy

### Status and Presentation

- extend [internal/proxy/status.go](/Users/sean/Programs/Clipal/internal/proxy/status.go)
- extend [internal/proxy/presentation.go](/Users/sean/Programs/Clipal/internal/proxy/presentation.go)
- update Web API/view types if surfaced in UI

## Testing Strategy

Add regression coverage for:

- strong sticky key extraction per capability
- `previous_response_id -> response.id` affinity recovery
- `L3` dynamic feature extraction for OpenAI, Anthropic, and Gemini message shapes
- no `L3` key when there is only one human message
- first-turn request with one human message still learns that human message on successful response completion
- request-side `second-to-last` human lookup matches response-side `last human` learning
- `L3` uses normalized text hash for actual lookup and 24-char preview for observability
- dynamic feature cache eviction by capacity
- response lookup assists `previous_response_id` chaining
- response completion writes response IDs and learned `L3` features into the correct caches
- changed or colliding `L3` feature falls back to normal cursor routing without error
- busy `429` does not deactivate provider
- busy `429` does not increment circuit failure
- buffered request body supports inline wait and retry without rereading `req.Body`
- concurrent busy updates merge by max window instead of overwriting
- request waits for `5s` busy window through fake clock or injected timing hooks
- retry succeeds on same provider and keeps binding
- retry fails again and overflows
- overflow rebinds later requests in same session
- busy provider does not get a stampede of immediate retries
- multi-key provider uses next key before cross-provider overflow
- all candidates busy returns nearest retry-after semantics
- hot reload preserves sticky/busy/response-lookup/L3 cache state according to config compatibility

## Risks

### 1. JSON Body Inspection Cost

Sticky extraction for request-side identifiers may require parsing buffered JSON bodies. Keep extractors narrow and avoid generic deep parsing.

### 2. Weak-Key Misbinding

Treat weak hints conservatively. Only strong keys get durable rebinding.

### 3. Busy Classification Ambiguity

Some providers blur rate limit and concurrency limit. Default to current cooldown behavior unless the signal is clearly short-lived busy.

### 4. Wait Amplification

Bound queueing to request context, enforce `max_inline_wait`, and use provider probe gating to avoid synchronized wakeups and unbounded user-visible stalls.

## Rollout Plan

Phase 1:

- add config and runtime structures
- add sticky extraction
- add busy state and classification
- implement request-local wait and overflow
- keep Web UI additions minimal

Phase 2:

- improve provider-specific busy heuristics
- expose richer Web observability
- tune defaults with real traffic feedback

## References

- OpenAI Responses API: https://platform.openai.com/docs/api-reference/responses/retrieve
- OpenAI tools/computer use guide: https://platform.openai.com/docs/guides/tools-computer-use
- Anthropic API getting started: https://docs.anthropic.com/en/api/getting-started
- Anthropic Messages examples: https://docs.anthropic.com/en/api/messages-examples
- Anthropic code execution tool and container reuse: https://docs.anthropic.com/en/docs/agents-and-tools/tool-use/code-execution-tool
- Anthropic prompt caching: https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching
- Gemini API docs: https://ai.google.dev/api
- Gemini context caching: https://ai.google.dev/gemini-api/docs/caching/
