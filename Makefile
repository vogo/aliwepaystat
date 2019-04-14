
static:
	go run makestatic/makestatic.go

build: static
	go build

