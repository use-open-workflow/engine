.PHONY: test build run fmt clean

APP_NAME := api
BUILD_DIR := bin
MAIN_PATH := ./cmd/api

test:
	go test -v ./...

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

fmt:
	go fmt ./...

clean:
	rm -rf $(BUILD_DIR)
