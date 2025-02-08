APP=devops
VERSION=1.0.0
GITHUB_OWNER=ginolatorilla
GITHUB_DOMAIN=github.com

COMMIT_HASH=$(shell git rev-parse HEAD)
PACKAGE=$(GITHUB_DOMAIN)/$(GITHUB_OWNER)/$(APP)

BUILD_FLAGS=-v -buildvcs
LD_FLAGS_COMMON=-X '$(PACKAGE)/cmd.AppName=$(APP)' -X '$(PACKAGE)/cmd.Version=$(VERSION)' -X '$(PACKAGE)/cmd.CommitHash=$(COMMIT_HASH)'
LD_FLAGS=-ldflags="$(LD_FLAGS_COMMON)"
LD_FLAGS_RELEASE=-ldflags="-s -w $(LD_FLAGS_COMMON)"
TEST_REGEX=".*"
TEST_PACKAGE="./..."
PREFIX?=$(HOME)/.local

.PHONY: all
all: test build

.PHONY: test
test: tidy
	@echo "🌡  Running tests..."
	go test -race $(BUILD_FLAGS) $(LD_FLAGS) -run $(TEST_REGEX) $(TEST_PACKAGE)

.PHONY: test/cover
test/cover: tidy
	@echo "🌡️  Running tests..."
	@go test -coverprofile=/tmp/coverage.out -race $(BUILD_FLAGS) $(LD_FLAGS) -run $(TEST_REGEX) $(TEST_PACKAGE)
	@go tool cover -html=/tmp/coverage.out

.PHONY: tidy
tidy:
	@echo "🧹 Tidying up package dependencies..."
	@go mod tidy

.PHONY: build
build:
	@echo "🏗️  Building the application..."
	go build $(BUILD_FLAGS) $(LD_FLAGS_RELEASE) -o bin/$(APP) $(PACKAGE)

.PHONY: install
install: test
	go install $(BUILD_FLAGS) $(LD_FLAGS_RELEASE) $(PACKAGE)
	mkdir -p $(PREFIX)/bin
	install scripts/* $(PREFIX)/bin
	ln -sf $(shell go env GOPATH)/bin/$(APP) $(PREFIX)/bin/kubectl-list_certs
	ln -sf $(shell go env GOPATH)/bin/$(APP) $(PREFIX)/bin/kubectl-lookup_address
	ln -sf $(shell go env GOPATH)/bin/$(APP) $(PREFIX)/bin/kubectl-list_unhealthy_pods
	ln -sf $(shell go env GOPATH)/bin/$(APP) $(PREFIX)/bin/kubectl-trigger_cronjob
	ln -sf $(shell go env GOPATH)/bin/$(APP) $(PREFIX)/bin/kubectl-list_addresses
	ln -sf $(shell go env GOPATH)/bin/$(APP) $(PREFIX)/bin/kubectl-list_finalizers

.PHONY: release
release: clean
	@$(MAKE) release-target GOOS=darwin GOARCH=arm64
	@$(MAKE) release-target GOOS=darwin GOARCH=amd64
	@$(MAKE) release-target GOOS=linux GOARCH=arm64
	@$(MAKE) release-target GOOS=linux GOARCH=amd64
	@$(MAKE) release-target GOOS=windows GOARCH=arm64
	@$(MAKE) release-target GOOS=windows GOARCH=amd64

.PHONY: release-target
release-target:
	@echo "🚀 Building release for $$GOOS/$$GOARCH..."
	@GOOS=$$GOOS GOARCH=$$GOARCH go build $(BUILD_FLAGS) $(LD_FLAGS_RELEASE) -o bin/$(APP) $(PACKAGE)
	@cp README.md LICENSE bin/
	@tar -czvf bin/$(APP)-$(VERSION)-$$GOOS-$$GOARCH.tar.gz -C bin $(APP) README.md LICENSE
	@sha256sum bin/$(APP)-$(VERSION)-$$GOOS-$$GOARCH.tar.gz > bin/checksum-$(APP)-$(VERSION)-$$GOOS-$$GOARCH.sha256.txt
	@echo "👍 Release artifacts created: "
	@echo "  📦 Tarball:          bin/$(APP)-$(VERSION)-$$GOOS-$$GOARCH.tar.gz"
	@echo "  📜 SHA-256 checksum: bin/checksum-$(APP)-$(VERSION)-$$GOOS-$$GOARCH.sha256.txt"

.PHONY: clean
clean:
	go clean
	go clean -testcache
	rm -rf bin/*

.PHONY: doc
doc:
	@go install golang.org/x/pkgsite/cmd/pkgsite@latest
	@pkgsite -open


.PHONY: help
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  help       - Show this help message"
	@echo "  all        - Run test, tidy, and build (default)"
	@echo "  install    - Installs the scripts in \$$PREFIX/bin (default is ~/.local/bin)"
	@echo "  test       - Run tests"
	@echo "  test/cover - Run tests with coverage"
	@echo "  tidy       - Sort out package dependencies"
	@echo "  build      - Build the application"
	@echo "  clean      - Clean up the build artifacts"
	@echo "  doc        - Open the documentation in the browser"
