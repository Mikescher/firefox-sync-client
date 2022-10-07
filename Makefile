
build:
	CGO_ENABLED=0 go build -o _out/ffsclient cmd/ffsclient/main.go
 
run: build
	./_out/ffsclient

clean:
	go clean
	rm ./_out/*