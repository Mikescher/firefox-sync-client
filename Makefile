
build:
	go build -o _out/ffsclient cmd/ffsclient/main.go
 
run:
	go build -o _out/ffsclient cmd/ffsclient/main.go
	./_out/ffsclient

clean:
	go clean
	rm ./_out/*