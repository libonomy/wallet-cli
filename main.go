//wallet-cli is a golang implementation of the libonomy node.
//See - https://libonomy.io
package main

import (
	"flag"
	"os"
	"syscall"

	"github.com/libonomy/wallet-cli/client"
	"github.com/libonomy/wallet-cli/repl"
	"github.com/libonomy/wallet-cli/wallet/accounts"
)

type mockClient struct {
}

func (m mockClient) LocalAccount() *accounts.Account {
	return nil
}

func (m mockClient) AccountInfo(id string) {

}
func (m mockClient) Transfer(from, to, amount, passphrase string) error {
	return nil
}

func main() {
	serverHostPort := client.DefaultNodeHostPort
	datadir := Getwd()

	flag.StringVar(&serverHostPort, "server", serverHostPort, "host:port of the libonomy node HTTP server")
	flag.StringVar(&datadir, "datadir", datadir, "The directory to store the wallet data within")
	flag.Parse()

	_, err := syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		return
	}
	be, err := client.NewWalletBE(serverHostPort, datadir)
	if err != nil {
		return
	}
	repl.Start(be)
}

func Getwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}
