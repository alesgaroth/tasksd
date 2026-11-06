.PHONY: all build test coverage validate hook run clean

BIN_DIR := bin
BINARY := $(BIN_DIR)/tasksd

all: test build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY) ./cmd/tasksd

test:
	go test -race -cover ./...
	@go test -coverprofile=core_cover.out ./internal/core > /dev/null
	@COVERAGE=$$(go tool cover -func=core_cover.out | grep total: | awk '{print $$3}' | sed 's/%//'); \
	rm -f core_cover.out; \
	echo "Core coverage: $$COVERAGE%"; \
	if [ "$$COVERAGE" != "100.0" ]; then \
		echo "Coverage on internal/core must be 100.0%, got $$COVERAGE%"; \
		exit 1; \
	fi

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

validate: build
	./$(BINARY) validate .

hook: build
	@cp scripts/pre-commit .git/hooks/pre-commit 2>/dev/null && chmod +x .git/hooks/pre-commit && echo "Pre-commit hook installed to .git/hooks/pre-commit" || echo "To enable the hook outside sandbox, run: cp scripts/pre-commit .git/hooks/pre-commit && chmod +x .git/hooks/pre-commit"

run: build
	./$(BINARY)

clean:
	rm -rf $(BIN_DIR) coverage.out core_cover.out
