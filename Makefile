format:
		goimports -w -l .
		go fmt

check:
		golangci-lint run --disable=unused,deadcode

test:
		go test

static:
	go run makestatic/makestatic.go

build: format check test static
	go build cmd/aliwepaystat/aliwepaystat.go

install: format check test static
	go install cmd/aliwepaystat/aliwepaystat.go

