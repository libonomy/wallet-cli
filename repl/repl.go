package repl

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/libonomy/ed25519"
	"github.com/libonomy/wallet-cli/accounts"
	"github.com/libonomy/wallet-cli/client"
	"github.com/libonomy/wallet-cli/log"
	"github.com/libonomy/wallet-cli/wallet/address"

	"github.com/c-bata/go-prompt"
)

const (
	prefix      = "$ "
	printPrefix = ">"
)

// TestMode variable used for check if unit test is running
var TestMode = false

type command struct {
	text        string
	description string
	fn          func()
}

type repl struct {
	commands []command
	client   Client
	input    string
}

// Client interface to REPL clients.
type Client interface {
	CreateAccount(alias string) *accounts.Account
	CurrentAccount() *accounts.Account
	SetCurrentAccount(a *accounts.Account)
	AccountInfo(address string) (*accounts.AccountInfo, error)
	NodeInfo() (*client.NodeInfo, error)
	Sanity() error
	Transfer(recipient address.Address, nonce, amount, gasPrice, gasLimit uint64, key ed25519.PrivateKey) (string, error)
	ListAccounts() []string
	GetAccount(name string) (*accounts.Account, error)
	StoreAccounts() error
	NodeURL() string
	Rebel(datadir string, space uint, coinbase string) error
	ListTxs(address string) ([]string, error)
	SetCoinbase(coinbase string) error

	//Unlock(passphrase string) error
	//IsAccountUnLock(id string) bool
	//Lock(passphrase string) error
	//SetVariables(params, flags []string) error
	//GetVariable(key string) string
	//Restart(params, flags []string) error
	//NeedRestartNode(params, flags []string) bool
	//Setup(allocation string) error
}

// Start starts REPL.
func Start(c Client) {
	if !TestMode {
		r := &repl{client: c}
		r.initializeCommands()

		runPrompt(r.executor, r.completer, r.firstTime, uint16(len(r.commands)))
	} else {
		// holds for unit test purposes
		hold := make(chan bool)
		<-hold
	}
}

func (r *repl) initializeCommands() {
	r.commands = []command{
		{"create-account", "Create a new account (key pair) and set as current", r.createAccount},
		{"use-previous", "Set one of the previously created accounts as current", r.chooseAccount},
		{"info", "Display the current account info", r.accountInfo},
		{"status", "Display the node status", r.nodeInfo},
		{"sign", "Sign a hex message with the current account private key", r.sign},
		{"textsign", "Sign a text message with the current account private key", r.textsign},
		{"quit", "Quit the CLI", r.quit},
	}
}

func (r *repl) executor(text string) {
	for _, c := range r.commands {
		if len(text) >= len(c.text) && text[:len(c.text)] == c.text {
			r.input = text
			//log.Debug(userExecutingCommandMsg, c.text)
			c.fn()
			return
		}
	}

	fmt.Println(printPrefix, "invalid command.")
}

func (r *repl) completer(in prompt.Document) []prompt.Suggest {
	suggets := make([]prompt.Suggest, 0)
	for _, command := range r.commands {
		s := prompt.Suggest{
			Text:        command.text,
			Description: command.description,
		}

		suggets = append(suggets, s)
	}

	return prompt.FilterHasPrefix(suggets, in.GetWordBeforeCursor(), true)
}

func (r *repl) firstTime() {

	if err := r.client.Sanity(); err != nil {
		log.Error("Failed to connect to node at %v: %v", r.client.NodeURL(), err)
		r.quit()
	}

	fmt.Println("Welcome to libonomy. Connected to node at ", r.client.NodeURL())
}

func (r *repl) chooseAccount() {
	accs := r.client.ListAccounts()
	if len(accs) == 0 {
		r.createAccount()
		return
	}

	fmt.Println(printPrefix, "Choose an account to load:")
	accName := multipleChoice(accs)
	account, err := r.client.GetAccount(accName)
	if err != nil {
		panic("wtf")
	}
	fmt.Printf("%s Loaded account alias: `%s`, address: %s \n", printPrefix, account.Name, accounts.StringAddress(account.Address()))

	r.client.SetCurrentAccount(account)
}

func (r *repl) createAccount() {
	fmt.Println(printPrefix, "Create a new account")
	alias := inputNotBlank(createAccountMsg)

	ac := r.client.CreateAccount(alias)
	err := r.client.StoreAccounts()
	if err != nil {
		log.Error("failed to create account: %v", err)
		return
	}

	fmt.Printf("%s Created account alias: `%s`, address: %s \n", printPrefix, ac.Name, accounts.StringAddress(ac.Address()))
	r.client.SetCurrentAccount(ac)
}

func (r *repl) commandLineParams(idx int, input string) string {
	c := r.commands[idx]
	params := strings.Replace(input, c.text, "", -1)

	return strings.TrimSpace(params)
}

func (r *repl) accountInfo() {
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	address := address.BytesToAddress(acc.PubKey)

	info, err := r.client.AccountInfo(hex.EncodeToString(address.Bytes()))
	if err != nil {
		log.Error("failed to get account info: %v", err)
		info = &accounts.AccountInfo{}
	}

	fmt.Println(printPrefix, "Local alias: ", acc.Name)
	fmt.Println(printPrefix, "Address: ", accounts.StringAddress(address))
	fmt.Println(printPrefix, "Balance: ", info.Balance)
	fmt.Println(printPrefix, "Nonce: ", info.Nonce)
	fmt.Println(printPrefix, fmt.Sprintf("Public key: 0x%s", hex.EncodeToString(acc.PubKey)))
	fmt.Println(printPrefix, fmt.Sprintf("Private key: 0x%s", hex.EncodeToString(acc.PrivKey)))
}

func (r *repl) nodeInfo() {
	info, err := r.client.NodeInfo()
	if err != nil {
		log.Error("failed to get node info: %v", err)
		return
	}

	fmt.Println(printPrefix, "Synced:", info.Synced)
	fmt.Println(printPrefix, "Synced layer:", info.SyncedLayer)
	fmt.Println(printPrefix, "Current layer:", info.CurrentLayer)
	fmt.Println(printPrefix, "Verified layer:", info.VerifiedLayer)
	fmt.Println(printPrefix, "Peers:", info.Peers)
	fmt.Println(printPrefix, "Min peers:", info.MinPeers)
	fmt.Println(printPrefix, "Max peers:", info.MaxPeers)
	fmt.Println(printPrefix, "Libonomy data directory:", info.LibonomyDatadir)
	fmt.Println(printPrefix, "Libonomy status:", info.LibonomyStatus)
	fmt.Println(printPrefix, "Libonomy coinbase:", info.LibonomyCoinbase)
	fmt.Println(printPrefix, "Libonomy remaining bytes:", info.LibonomyRemainingBytes)
}

func (r *repl) transferCoins() {
	fmt.Println(printPrefix, initialTransferMsg)
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	srcAddress := address.BytesToAddress(acc.PubKey)
	info, err := r.client.AccountInfo(hex.EncodeToString(srcAddress.Bytes()))
	if err != nil {
		log.Error("failed to get account info: %v", err)
		return
	}

	destAddressStr := inputNotBlank(destAddressMsg)
	destAddress := address.HexToAddress(destAddressStr)

	amountStr := inputNotBlank(amountToTransferMsg)

	gas := uint64(1)
	if yesOrNoQuestion(useDefaultGasMsg) == "n" {
		gasStr := inputNotBlank(enterGasPrice)
		gas, err = strconv.ParseUint(gasStr, 10, 64)
		if err != nil {
			log.Error("invalid gas", err)
			return
		}
	}

	fmt.Println(printPrefix, "Transaction summary:")
	fmt.Println(printPrefix, "From:  ", srcAddress.String())
	fmt.Println(printPrefix, "To:    ", destAddress.String())
	fmt.Println(printPrefix, "Amount:", amountStr)
	fmt.Println(printPrefix, "Gas:   ", gas)
	fmt.Println(printPrefix, "Nonce: ", info.Nonce)

	nonce, err := strconv.ParseUint(info.Nonce, 10, 64)
	amount, err := strconv.ParseUint(amountStr, 10, 64)

	if yesOrNoQuestion(confirmTransactionMsg) == "y" {
		id, err := r.client.Transfer(destAddress, nonce, amount, gas, 100, acc.PrivKey)
		if err != nil {
			log.Error(err.Error())
			return
		}
		fmt.Println(printPrefix, fmt.Sprintf("tx submitted, id: %v", id))
	}
}

func (r *repl) rebel() {
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	datadir := inputNotBlank(libonomyDatadirMsg)

	spaceStr := inputNotBlank(libonomySpaceAllocationMsg)
	space, err := strconv.ParseUint(spaceStr, 10, 32)
	if err != nil {
		log.Error("failed to parse: %v", err)
		return
	}

	if err := r.client.Rebel(datadir, uint(space)<<30, accounts.StringAddress(acc.Address())); err != nil {
		log.Error("failed to start libonomy: %v", err)
		return
	}
}

func (r *repl) listTxs() {
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	txs, err := r.client.ListTxs(accounts.StringAddress(acc.Address()))
	if err != nil {
		log.Error("failed to list txs: %v", err)
		return
	}

	fmt.Println(printPrefix, fmt.Sprintf("txs: %v", txs))
}

func (r *repl) quit() {
	os.Exit(0)
}

func (r *repl) coinbase() {
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	if err := r.client.SetCoinbase(accounts.StringAddress(acc.Address())); err != nil {
		log.Error("failed to set coinbase: %v", err)
		return
	}
}

func (r *repl) sign() {
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	msgStr := inputNotBlank(msgSignMsg)
	msg, err := hex.DecodeString(msgStr)
	if err != nil {
		log.Error("failed to decode msg hex string: %v", err)
		return
	}

	signature := ed25519.Sign2(acc.PrivKey, msg)

	fmt.Println(printPrefix, fmt.Sprintf("signature (in hex): %x", signature))
}

func (r *repl) textsign() {
	acc := r.client.CurrentAccount()
	if acc == nil {
		r.chooseAccount()
		acc = r.client.CurrentAccount()
	}

	msg := inputNotBlank(msgTextSignMsg)
	signature := ed25519.Sign2(acc.PrivKey, []byte(msg))

	fmt.Println(printPrefix, fmt.Sprintf("signature (in hex): %x", signature))
}

/*
func (r *repl) unlockAccount() {
	passphrase := r.commandLineParams(1, r.input)
	err := r.client.Unlock(passphrase)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	acctCmd := r.commands[3]
	r.executor(fmt.Sprintf("%s %s", acctCmd.text, passphrase))
}

func (r *repl) lockAccount() {
	passphrase := r.commandLineParams(2, r.input)
	err := r.client.Lock(passphrase)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	acctCmd := r.commands[3]
	r.executor(fmt.Sprintf("%s %s", acctCmd.text, passphrase))
}*/
