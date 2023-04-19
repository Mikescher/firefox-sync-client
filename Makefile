
# ./firefox-sync-client
#########################

build:
	CGO_ENABLED=0 go build -o _out/ffsclient ./cmd/ffsclient

run: build
	./_out/ffsclient

clean:
	go clean
	rm -rf ./_out/*

package:
	go clean
	rm -rf ./_out/*

	GOARCH=386   GOOS=linux   CGO_ENABLED=0 go build -o _out/ffsclient_linux-386-static                                      ./cmd/ffsclient  # Linux - 32 bit
	GOARCH=amd64 GOOS=linux   CGO_ENABLED=0 go build -o _out/ffsclient_linux-amd64-static                                    ./cmd/ffsclient  # Linux - 64 bit
	GOARCH=arm64 GOOS=linux   CGO_ENABLED=0 go build -o _out/ffsclient_linux-arm64-static                                    ./cmd/ffsclient  # Linux - ARM
	GOARCH=386   GOOS=linux                 go build -o _out/ffsclient_linux-386                                             ./cmd/ffsclient  # Linux - 32 bit
	GOARCH=amd64 GOOS=linux                 go build -o _out/ffsclient_linux-amd64                                           ./cmd/ffsclient  # Linux - 64 bit
	GOARCH=arm64 GOOS=linux                 go build -o _out/ffsclient_linux-arm64                                           ./cmd/ffsclient  # Linux - ARM
	GOARCH=386   GOOS=windows               go build -o _out/ffsclient_win-386.exe         -tags timetzdata -ldflags "-w -s" ./cmd/ffsclient  # Windows - 32 bit
	GOARCH=amd64 GOOS=windows               go build -o _out/ffsclient_win-amd64.exe       -tags timetzdata -ldflags "-w -s" ./cmd/ffsclient  # Windows - 64 bit
	GOARCH=arm64 GOOS=windows               go build -o _out/ffsclient_win-arm64.exe       -tags timetzdata -ldflags "-w -s" ./cmd/ffsclient  # Windows - ARM
	GOARCH=amd64 GOOS=darwin                go build -o _out/ffsclient_macos-amd64                                           ./cmd/ffsclient  # macOS - 32 bit
	GOARCH=amd64 GOOS=darwin                go build -o _out/ffsclient_macos-amd64                                           ./cmd/ffsclient  # macOS - 64 bit
	GOARCH=amd64 GOOS=openbsd               go build -o _out/ffsclient_openbsd-amd64                                         ./cmd/ffsclient  # OpenBSD - 64 bit
	GOARCH=arm64 GOOS=openbsd               go build -o _out/ffsclient_openbsd-arm64                                         ./cmd/ffsclient  # OpenBSD - ARM
	GOARCH=amd64 GOOS=freebsd               go build -o _out/ffsclient_freebsd-amd64                                         ./cmd/ffsclient  # FreeBSD - 64 bit
	GOARCH=arm64 GOOS=freebsd               go build -o _out/ffsclient_freebsd-arm64                                         ./cmd/ffsclient  # FreeBSD - ARM

	_data/package-data/deb.sh
	_data/package-data/aur-git.sh
	_data/package-data/aur-bin.sh
	_data/package-data/chocolatey.sh
	_data/package-data/homebrew.sh

	echo ""
	echo "[TODO]: call 'make package-push-aur-git'    "
	echo "[TODO]: call 'make package-push-aur-bin'    "
	echo "[TODO]: call 'make package-push-homebrew'   "
	echo "[TODO]: call 'make package-push-chocolatey' "
	echo "[TODO]: create github release"
	echo ""

package-push-aur-git:
	cd _out/ffsclient-git && git push

package-push-aur-bin:
	cd _out/ffsclient-bin && git push

package-push-homebrew:
	cd _out/homebrew-tap && git push

package-push-chocolatey:
	docker run --rm --volume "$(shell pwd)/_out/chocolatey:/root/ffsclient" "chocolatey/choco:latest" /root/ffsclient/push.sh "$(shell secret-tool lookup Path "Personal/References/ChocoAPIKey")"
