VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE     := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
  -X main.Version=$(VERSION) \
  -X main.Commit=$(COMMIT) \
  -X main.Date=$(DATE)

.PHONY: build test lint clean

build:
	go build -ldflags "$(LDFLAGS)" -o xin-code .

test:
	go test -v -cover -race -timeout=60s ./...

lint:
	golangci-lint run

clean:
	rm -f xin-code

install: build
	cp xin-code $(GOPATH)/bin/xin-code
