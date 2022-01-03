SOURCE_FILES := $(shell find . -iname '*.go' ! -ipath '*/vendor/*')

metie: $(SOURCE_FILES)
	go build .

build-linux: $(SOURCE_FILES)
	env GOOS=linux GOARCH=amd64 go build -o metie-linux .

test:
	go test -v ./...

lint:
	go vet ./...
	golangci-lint run

.PHONY: test lint build-arm
