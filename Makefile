BINARY  := arcane
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/bensch98/arcane/cmd.Version=$(VERSION)
GOFLAGS := -trimpath -ldflags '$(LDFLAGS)'

.PHONY: all build test clean install

all: build

build:
	go build $(GOFLAGS) -o $(BINARY) .

test:
	go test ./...

clean:
	rm -f $(BINARY)

install:
	go install $(GOFLAGS) .
