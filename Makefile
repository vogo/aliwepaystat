version := v1.3.0

format:
		goimports -w -l .
		go fmt

check:
		golangci-lint run --disable=unused,deadcode

test:
		go test

build: format check test
	rm -f dist/*.zip
	cd dist && GOOS=linux go build ../cmd/aliwepaystat/aliwepaystat.go && zip aliwepaystat-$(version)-linux.zip aliwepaystat && rm -f aliwepaystat
	cd dist && GOOS=darwin go build ../cmd/aliwepaystat/aliwepaystat.go && zip aliwepaystat-$(version)-mac.zip aliwepaystat && rm -f aliwepaystat
	cd dist && GOOS=windows go build ../cmd/aliwepaystat/aliwepaystat.go && zip aliwepaystat-$(version)-windows.zip aliwepaystat.exe && rm -f aliwepaystat.exe

install: format check test
	go install cmd/aliwepaystat/aliwepaystat.go

