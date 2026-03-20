version := v1.3.0

format:
		goimports -w -l .
		go fix ./...
		go fmt

check:
		golangci-lint run

test:
		go test -coverprofile=coverage.out -covermode=atomic -v

build: format check test

install: format check test
	go install cmd/aliwepaystat/aliwepaystat.go

