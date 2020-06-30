BINARY := cli_wallet
LINUX=$(BINARY)_linux_amd64


build-linux:
	env GOOS=linux GOARCH=amd64 go build -o $(LINUX)

