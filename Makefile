APP=devops
VERSION=0.1.0
GITHUB_OWNER=ginolatorilla
GITHUB_DOMAIN=github.com

COMMIT_HASH=$(shell git rev-parse HEAD)
PACKAGE=$(GITHUB_DOMAIN)/$(GITHUB_OWNER)/$(APP)

BUILD_FLAGS=-v -buildvcs 
LD_FLAGS=-ldflags="-X '$(PACKAGE)/cmd.AppName=$(APP)' -X '$(PACKAGE)/cmd.Version=$(VERSION)' -X '$(PACKAGE)/cmd.CommitHash=$(COMMIT_HASH)'"
TEST_REGEX=".*"
TEST_PACKAGE="./..."

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
	go mod tidy

.PHONY: build
build:
	@echo "🏗️  Building the application..."
	go build $(BUILD_FLAGS) $(LD_FLAGS) -o bin/$(APP) $(PACKAGE) 

.PHONY: clean
clean:
	go clean
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
	@echo "  test       - Run tests"
	@echo "  test/cover - Run tests with coverage"
	@echo "  tidy       - Sort out package dependencies"
	@echo "  build      - Build the application"
	@echo "  clean      - Clean up the build artifacts"
	@echo "  doc        - Open the documentation in the browser"
