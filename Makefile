.PHONY: build run clean install

BINARY := oli
BUILD_DIR := ./bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/ollama-cli

run: build
	$(BUILD_DIR)/$(BINARY) $(ARGS)

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/
