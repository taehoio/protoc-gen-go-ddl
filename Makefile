GOPATH:=$(shell go env GOPATH)
APP?=protoc-gen-go-ddl

.PHONY: build
## build: build the application
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o bin/${APP}.linux.amd64 cmd/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -ldflags="-w -s" -o bin/${APP}.linux.arm64 cmd/main.go
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-w -s" -o bin/${APP} cmd/main.go

.PHONY: run
## run: run the application
run:
	go run -v -race cmd/main.go

.PHONY: format
## format: format files
format:
	@go install golang.org/x/tools/cmd/goimports@latest
	goimports -local github.com/taehoio -w .
	gofmt -s -w .
	go mod tidy

.PHONY: test
## test: run tests
test:
	@go install github.com/rakyll/gotest@latest
	gotest -p 1 -race -cover -v ./...

.PHONY: coverage
## coverage: run tests with coverage
coverage:
	@go install github.com/rakyll/gotest@latest
	gotest -p 1 -race -coverprofile=coverage.out -covermode=atomic -v ./...

.PHONY: lint
## lint: check everything's okay
lint:
	@go install github.com/kyoh86/scopelint@latest
	golangci-lint run ./...
	scopelint --set-exit-status ./...
	go mod verify

.PHONY: generate
## generate: generate source code for mocking
generate:
	@go install golang.org/x/tools/cmd/stringer@latest
	@go install github.com/golang/mock/mockgen@v1.6.0
	go generate ./...

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':'
