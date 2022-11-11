
# ./firefox-sync-client
#########################

build:
	CGO_ENABLED=0 go build -o _out/ffsclient cmd/ffsclient/main.go

run: build
	./_out/ffsclient

clean:
	go clean
	rm -rf ./_out/*

package:
	go clean
	rm -rf ./_out/*

	GOARCH=386   GOOS=linux   CGO_ENABLED=0 go build -o _out/ffsclient_linux-386-static                                      cmd/ffsclient/main.go  # Linux - 32 bit
	GOARCH=amd64 GOOS=linux   CGO_ENABLED=0 go build -o _out/ffsclient_linux-amd64-static                                    cmd/ffsclient/main.go  # Linux - 64 bit
	GOARCH=arm64 GOOS=linux   CGO_ENABLED=0 go build -o _out/ffsclient_linux-arm64-static                                    cmd/ffsclient/main.go  # Linux - ARM
	GOARCH=386   GOOS=linux                 go build -o _out/ffsclient_linux-386                                             cmd/ffsclient/main.go  # Linux - 32 bit
	GOARCH=amd64 GOOS=linux                 go build -o _out/ffsclient_linux-amd64                                           cmd/ffsclient/main.go  # Linux - 64 bit
	GOARCH=arm64 GOOS=linux                 go build -o _out/ffsclient_linux-arm64                                           cmd/ffsclient/main.go  # Linux - ARM
	GOARCH=386   GOOS=windows               go build -o _out/ffsclient_win-386.exe         -tags timetzdata -ldflags "-w -s" cmd/ffsclient/main.go  # Windows - 32 bit
	GOARCH=amd64 GOOS=windows               go build -o _out/ffsclient_win-amd64.exe       -tags timetzdata -ldflags "-w -s" cmd/ffsclient/main.go  # Windows - 64 bit
	GOARCH=arm64 GOOS=windows               go build -o _out/ffsclient_win-arm64.exe       -tags timetzdata -ldflags "-w -s" cmd/ffsclient/main.go  # Windows - ARM
	GOARCH=amd64 GOOS=darwin                go build -o _out/ffsclient_macos-amd64                                           cmd/ffsclient/main.go  # macOS - 32 bit
	GOARCH=amd64 GOOS=darwin                go build -o _out/ffsclient_macos-amd64                                           cmd/ffsclient/main.go  # macOS - 64 bit
	GOARCH=amd64 GOOS=openbsd               go build -o _out/ffsclient_openbsd-amd64                                         cmd/ffsclient/main.go  # OpenBSD - 64 bit
	GOARCH=arm64 GOOS=openbsd               go build -o _out/ffsclient_openbsd-arm64                                         cmd/ffsclient/main.go  # OpenBSD - ARM
	GOARCH=amd64 GOOS=freebsd               go build -o _out/ffsclient_freebsd-amd64                                         cmd/ffsclient/main.go  # FreeBSD - 64 bit
	GOARCH=arm64 GOOS=freebsd               go build -o _out/ffsclient_freebsd-arm64                                         cmd/ffsclient/main.go  # FreeBSD - ARM

	_data/package-data/deb.sh
	_data/package-data/aur-git.sh
	_data/package-data/aur-bin.sh
	_data/package-data/chocolatey.sh
	_data/package-data/homebrew.sh
