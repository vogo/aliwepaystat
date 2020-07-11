version := 1.0

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
	cd dist && GOOS=linux go build ../cmd/aliwepaystat/aliwepaystat.go && zip aliwepaystat-$(version)-linux.zip aliwepaystat && rm -f aliwepaystat
	cd dist && GOOS=darwin go build ../cmd/aliwepaystat/aliwepaystat.go && zip aliwepaystat-$(version)-mac.zip aliwepaystat && rm -f aliwepaystat
	cd dist && GOOS=windows go build ../cmd/aliwepaystat/aliwepaystat.go && zip aliwepaystat-$(version)-windows.zip aliwepaystat.exe && rm -f aliwepaystat.exe

install: format check test static
	go install cmd/aliwepaystat/aliwepaystat.go

