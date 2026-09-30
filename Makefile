BIN_DIR := bin

.PHONY: build test fmt clean

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/mach ./cli/cmd/mach
	go build -o $(BIN_DIR)/mach-agent ./agent/cmd/mach-agent
	go build -o $(BIN_DIR)/mach-controller ./controller/cmd/mach-controller

test:
	go test ./...

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')

clean:
	rm -rf $(BIN_DIR)
