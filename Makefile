.PHONY: build test test-unit test-smoke test-live-oauth test-oauth-authorize lint vuln fmt ci

GO ?= go
GOBIN := $(shell $(GO) env GOPATH)/bin
GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null || echo "$(GOBIN)/golangci-lint")
GOVULNCHECK ?= $(shell command -v govulncheck 2>/dev/null || echo "$(GOBIN)/govulncheck")

build:
	./scripts/build.sh

test: test-unit

test-unit:
	$(GO) test ./...

test-smoke:
	./scripts/smoke_test.sh

test-live-oauth:
	CLIPAL_LIVE_CONFIG_DIR="$(CONFIG_DIR)" \
	CLIPAL_LIVE_OAUTH_REF="$(OAUTH_REF)" \
	CLIPAL_LIVE_OAUTH_FILE="$(OAUTH_FILE)" \
	CLIPAL_LIVE_MODEL="$(MODEL)" \
	CLIPAL_LIVE_SKIP_STREAM="$(SKIP_STREAM)" \
	CLIPAL_LIVE_SKIP_REFRESH_RETRY="$(SKIP_REFRESH_RETRY)" \
	CLIPAL_LIVE_KEEP_TEMP="$(KEEP_TEMP)" \
	./scripts/live_oauth_smoke.sh

test-oauth-authorize:
	./scripts/oauth_authorize_smoke.sh --mock

lint:
	$(GOLANGCI_LINT) run ./...

vuln:
	$(GOVULNCHECK) ./...

fmt:
	gofmt -w $$(find cmd internal -type f -name '*.go' | sort)

ci: test-unit lint vuln test-smoke
