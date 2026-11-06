.PHONY: all build test coverage run clean

BIN_DIR := bin
BINARY := $(BIN_DIR)/tasksd

all: test build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY) ./cmd/tasksd

test:
	go test -race -cover ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

run: build
	./$(BINARY)

clean:
	rm -rf $(BIN_DIR) coverage.out
