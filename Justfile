build:
	go mod tidy
	go build -o label-logger-go main.go

install:
	just build
	cp label-logger-go ~/.local/bin
