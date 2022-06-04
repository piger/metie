SOURCE_FILES := $(shell find . -iname '*.go' ! -ipath '*/vendor/*')

metie: $(SOURCE_FILES)
	go build .

build-linux: $(SOURCE_FILES)
	env GOOS=linux GOARCH=amd64 go build -o metie-linux .
.PHONY: build-linux

test:
	go test -v ./...
.PHONY: test

lint:
	go vet ./...
	golangci-lint run
.PHONY: lint

release:
	goreleaser --rm-dist
.PHONY: release
