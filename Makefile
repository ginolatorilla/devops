APP=devops
VERSION=1.0.1
GITHUB_OWNER=ginolatorilla
GITHUB_DOMAIN=github.com

COMMIT_HASH=$(shell git rev-parse HEAD)
PACKAGE=$(GITHUB_DOMAIN)/$(GITHUB_OWNER)/$(APP)

BUILD_FLAGS=-v -buildvcs
LD_FLAGS_COMMON=-X '$(PACKAGE)/cmd/devops/cmd.AppName=$(APP)' -X '$(PACKAGE)/cmd/devops/cmd.Version=$(VERSION)' -X '$(PACKAGE)/cmd/devops/cmd.CommitHash=$(COMMIT_HASH)'
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
	@echo "🏗️  Building devops-cli..."
	go build $(BUILD_FLAGS) $(LD_FLAGS_RELEASE) -o bin/ $(PACKAGE)/cmd/...

.PHONY: install
install: test
	go install $(BUILD_FLAGS) $(LD_FLAGS_RELEASE) $(PACKAGE)/cmd/...
	mkdir -p $(PREFIX)/bin
	install scripts/* $(PREFIX)/bin

.PHONY: release
release: clean
	@for os in darwin linux windows; do \
		for arch in amd64 arm64; do \
			$(MAKE) release-target GOOS=$$os GOARCH=$$arch; \
		done; \
	done

.PHONY: release-target
release-target:
	@echo "🚀 Building release for $$GOOS/$$GOARCH..."
	@mkdir -p bin/release-$$GOOS-$$GOARCH
	@cp LICENSE README.md bin/release-$$GOOS-$$GOARCH
	@cp -r scripts bin/release-$$GOOS-$$GOARCH/scripts
	@GOOS=$$GOOS GOARCH=$$GOARCH go build $(BUILD_FLAGS) $(LD_FLAGS_RELEASE) -o bin/release-$$GOOS-$$GOARCH $(PACKAGE)/cmd/...
	@cd bin/release-$$GOOS-$$GOARCH && tar -czvf ../$(APP)-$(VERSION)-$$GOOS-$$GOARCH.tar.gz *
	@sha256sum bin/$(APP)-$(VERSION)-$$GOOS-$$GOARCH.tar.gz > bin/checksum-$(VERSION)-$$GOOS-$$GOARCH.sha256.txt
	@echo "👍 Release artifacts created for $$plugin: "
	@echo "  📦 Tarball:          bin/$(APP)-$(VERSION)-$$GOOS-$$GOARCH.tar.gz"
	@echo "  📜 SHA-256 checksum: bin/checksum-$(VERSION)-$$GOOS-$$GOARCH.sha256.txt"

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
