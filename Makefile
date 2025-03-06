BINARY_NAME=shortener

BUILD_DIR=cmd/shortener

# Get git commit hash
COMMIT_HASH=$(shell git rev-parse --short HEAD)
BUILD_DATE=$(shell date -u +"%Y-%m-%d_%H:%M:%S")
VERSION=1.0.0

all: build

build:
	cd $(BUILD_DIR) && go build -ldflags "-X main.buildVersion=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.buildCommit=$(COMMIT_HASH)" -o $(BINARY_NAME)

run:
	go run $(BUILD_DIR)/main.go

clean:
	cd $(BUILD_DIR) && rm -f $(BINARY_NAME)
