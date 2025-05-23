# Define root directory, binary name, and Go files
root_dir := $(shell pwd)
BIN_NAME := $(root_dir)/otp_dev
GO_FILES := $(root_dir)/sms.go  

.build: build

build:
	GOARCH=amd64 GOOS=linux go build -o $(BIN_NAME) $(GO_FILES)

.run: run

run:
	go run $(GO_FILES)

.install: install

install: build
	sudo mv $(BIN_NAME) /usr/local/bin/$(otp_dev)

.clean: clean

clean:
	rm -f $(BIN_NAME)

help:
	@echo "makefile_commands:"
	@echo "  make build   - Build the Go binary"
	@echo "  make run     - Run the Go program"
	@echo "  make install - Install the binary to /usr/local/bin"
	@echo "  make clean   - Clean up the binary file"
