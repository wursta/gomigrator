BIN := "./bin/gomigrator"
GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X github.com/wursta/gomigrator/cmd.release="develop" -X github.com/wursta/gomigrator/cmd.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X github.com/wursta/gomigrator/cmd.gitHash=$(GIT_HASH)

build-win:
	GOOS=windows GOARCH=amd64 go build -v -o $(BIN).exe -ldflags "$(LDFLAGS)" ./
install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2
lint: install-lint-deps
	golangci-lint run ./...
lint-fix: install-lint-deps
	golangci-lint run --fix ./...
test:
	go test -race ./internal/...
