SOURCE_FILES := $(shell find . -iname '*.go' ! -ipath '*/vendor/*')

metie: $(SOURCE_FILES)
	go build .

build-linux: $(SOURCE_FILES)
	env GOOS=linux GOARCH=amd64 go build -o metie-linux .
.PHONY: build-linux

build-arm: $(SOURCE_FILES)
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build -o metie-arm .
.PHONY: build-arm

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
