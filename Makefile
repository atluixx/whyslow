VERSION ?= dev
BINARY ?= build/whyslow

.PHONY: build test install uninstall clean

build:
	@mkdir -p build
	go build -ldflags "-X main.Version=$(VERSION)" -o $(BINARY) ./cmd/whyslow

test:
	go test ./...
	go vet ./...

install: build
	$(BINARY) install

uninstall: build
	$(BINARY) uninstall

clean:
	rm -rf build
