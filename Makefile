BINARY  := arcane
GOFLAGS := -trimpath

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
