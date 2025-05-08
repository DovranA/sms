# Define root directory, binary name, and Go files
root_dir := $(shell pwd)
BIN_NAME := $(root_dir)/otp_dev
GO_FILES := $(root_dir)/sms.go  

# Go build command to build the binary
build:
  GOARCH=amd64 GOOS=linux go build -o $(BIN_NAME) $(GO_FILES)

# Go run command to run the program

run:
  go run $(GO_FILES)

# Install command (useful for installing Go binaries globally)
install: build
  sudo mv $(BIN_NAME) /usr/local/bin/$(BIN_NAME)

# Clean up the binary file
clean:
  rm -f $(BIN_NAME)

# Help command to print available commands
help:
  @echo "makefile_commands:"
  @echo "  make build   - Build the Go binary"
  @echo "  make run     - Run the Go program"
  @echo "  make install - Install the binary to /usr/local/bin"
  @echo "  make clean   - Clean up the binary file"