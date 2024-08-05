# Build
FROM golang:1.22-alpine as builder

RUN mkdir -p /opt/migrator

WORKDIR /opt/migrator

COPY cmd cmd
COPY internal internal
COPY pkg pkg
COPY main.go main.go
COPY go.mod go.mod
COPY go.sum go.sum

RUN CGO_ENABLED=0 GOOS=linux go build -v -o ./bin/gomigrator .

# Test
FROM golang:1.22-alpine

WORKDIR /root/

COPY go.mod go.mod
COPY go.sum go.sum
COPY ./intergation_tests intergation_tests

RUN mkdir -p intergation_tests/bin
COPY --from=builder /opt/migrator/bin/gomigrator /root/intergation_tests/bin

CMD GOMIGRATOR_TEST_BINARY=/root/intergation_tests/bin/gomigrator go test ./intergation_tests/...