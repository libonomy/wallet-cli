package accounts

import (
	"encoding/hex"
	"fmt"

	"github.com/libonomy/ed25519"
	"github.com/libonomy/wallet-cli/wallet/address"
)

type Account struct {
	Name    string
	PrivKey ed25519.PrivateKey // the pub & private key
	PubKey  ed25519.PublicKey  // only the pub key part
}

func (a *Account) Address() address.Address {
	return address.BytesToAddress(a.PubKey[:])
}

func StringAddress(addr address.Address) string {
	return fmt.Sprintf("0x%s", hex.EncodeToString(addr.Bytes()))
}

type AccountInfo struct {
	Nonce   string
	Balance string
}

func (s Store) GetAccount(name string) (*Account, error) {
	if acc, ok := s[name]; ok {
		priv, err := hex.DecodeString(acc.PrivKey)
		if err != nil {
			return nil, err
		}
		pub, err := hex.DecodeString(acc.PubKey)
		if err != nil {
			return nil, err
		}

		return &Account{name, priv, pub}, nil
	}
	return nil, fmt.Errorf("account not found")
}

func (s Store) ListAccounts() []string {
	lst := make([]string, 0, len(s))
	for key := range s {
		lst = append(lst, key)
	}
	return lst
}
