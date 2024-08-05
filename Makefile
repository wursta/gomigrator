BIN := "./bin/gomigrator"
INTEGRATION_TEST_BIN = "./intergation_tests/bin/gomigrator"
GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X github.com/wursta/gomigrator/cmd.release="develop" -X github.com/wursta/gomigrator/cmd.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X github.com/wursta/gomigrator/cmd.gitHash=$(GIT_HASH)

build:
	CGO_ENABLED=0 GOOS=linux go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./
build-for-integration-test:
	CGO_ENABLED=0 GOOS=linux go build -v -o $(INTEGRATION_TEST_BIN) -ldflags "$(LDFLAGS)" ./
build-win:
	GOOS=windows GOARCH=amd64 go build -v -o $(BIN).exe -ldflags "$(LDFLAGS)" ./
build-win-for-integration-test:
	GOOS=windows GOARCH=amd64 go build -v -o $(INTEGRATION_TEST_BIN).exe -ldflags "$(LDFLAGS)" ./
install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.59.1
lint: install-lint-deps
	golangci-lint run ./... --timeout=10m
lint-fix: install-lint-deps
	golangci-lint run --fix ./... --timeout=10m
test:
	go test -race -count=100 ./internal/...
run-integration-test:
	go test ./intergation_tests/...
build-and-run-integration-test: build-for-integration-test run-integration-test
integration-test: 
	docker compose -f ./deployments/docker-compose.integration-test.yaml up --build --abort-on-container-exit

# docker
build-docker:
	docker compose -f ./deployments/docker-compose.yaml run --rm -it migrator make build
build-win-docker:
	docker compose -f ./deployments/docker-compose.yaml run --rm -it migrator make build-win
up-docker:
	docker compose -f deployments/docker-compose.yaml up -d --build
down-docker:
	docker compose -f deployments/docker-compose.yaml down
lint-fix-docker:
	docker compose -f ./deployments/docker-compose.yaml run --rm -it migrator make lint-fix
lint-docker:
	docker compose -f ./deployments/docker-compose.yaml run --rm -it migrator make lint
test-docker:
	docker compose -f ./deployments/docker-compose.yaml run --rm -it migrator make test
integration-test-docker:
	docker compose -f ./deployments/docker-compose.yaml run --rm -it migrator make integration-test


