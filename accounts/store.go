package accounts

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/libonomy/ed25519"
	"github.com/libonomy/wallet-cli/os/log"
)

type AccountKeys struct {
	PubKey  string `json:"pubkey"`
	PrivKey string `json:"privkey"`
}

type Store map[string]AccountKeys

func StoreAccounts(path string, store *Store) error {
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	defer w.Close()
	if err := enc.Encode(store); err != nil {
		return err
	}
	return nil
}

func LoadAccounts(path string) (*Store, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Warning("genesis config not lodad since file does not exist. file=%v", path)
		return nil, err
	}
	r, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer r.Close()

	dec := json.NewDecoder(r)
	cfg := &Store{}
	err = dec.Decode(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (s Store) CreateAccount(alias string) *Account {
	sPub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Error("cannot create account: %s", err)
		return nil
	}
	acc := &Account{Name: alias, PubKey: sPub, PrivKey: key}
	s[alias] = AccountKeys{PubKey: hex.EncodeToString(sPub), PrivKey: hex.EncodeToString(key)}
	return acc
}
