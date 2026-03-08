LOCAL_BIN := $(CURDIR)/bin
export PATH := $(PATH):$(LOCAL_BIN)

GOLANGCI_LINT_VERSION := v2.11.2
BUF_VERSION := v1.66.0
BUF := $(LOCAL_BIN)/buf

# All targets that aren't real files must be declared .PHONY
# Without this: if a file named "test" exists, make test silently does nothing
.PHONY: bin-deps run test test-race lint generate proto-gen

# Install all dev tools locally — no global installs, no sudo required
bin-deps:
	GOBIN=$(LOCAL_BIN) go install go.uber.org/mock/mockgen@v0.6.0
	GOBIN=$(LOCAL_BIN) go install gotest.tools/gotestsum@v1.13.0
	GOBIN=$(LOCAL_BIN) curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(LOCAL_BIN) $(GOLANGCI_LINT_VERSION)
	GOBIN=$(LOCAL_BIN) go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)

run:
	go run cmd/main.go

# gotestsum: better output than plain go test, JUnit XML for CI
test:
	$(LOCAL_BIN)/gotestsum --format pkgname -- ./...

# Race detector: run in CI always, locally before every PR
test-race:
	$(LOCAL_BIN)/gotestsum --format pkgname -- -race ./...

lint:
	$(LOCAL_BIN)/golangci-lint run ./...

buff-dep:
	$(BUF) dep update

buff-lint:
	$(BUF) lint

buff-gen:
	$(BUF) dep prune
	$(BUF) build
	$(BUF) generate --template buf.gen.local.yaml
	$(BUF) generate --template buf.gen.external.yaml