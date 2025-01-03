BINARY_NAME=shortener

BUILD_DIR=cmd/shortener

all: build

build:
	cd $(BUILD_DIR) && go build -o $(BINARY_NAME)

run:
	go run $(BUILD_DIR)/main.go

clean:
	cd $(BUILD_DIR) && rm -f $(BINARY_NAME)
