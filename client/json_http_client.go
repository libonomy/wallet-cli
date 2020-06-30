package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/libonomy/wallet-cli/accounts"
	"github.com/libonomy/wallet-cli/os/log"
)

const DefaultNodeHostPort = "localhost:9090"

type Requester interface {
	Get(api, data string) []byte
}

type HTTPRequester struct {
	*http.Client
	url string
}

func NewHTTPRequester(url string) *HTTPRequester {
	return &HTTPRequester{&http.Client{}, url}
}

func (hr *HTTPRequester) Get(api, data string, logIO bool) (map[string]interface{}, error) {
	var jsonStr = []byte(data)
	url := hr.url + api
	if logIO {
		log.Info("request: %v, body: %v", url, data)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := hr.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, _ := ioutil.ReadAll(res.Body)
	if logIO {
		log.Info("response body: %s", resBody)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("`%v` response status code: %d", api, res.StatusCode)
	}

	var f interface{}
	err = json.NewDecoder(bytes.NewBuffer(resBody)).Decode(&f)
	if err != nil {
		return nil, err
	}

	return f.(map[string]interface{}), nil
}

type HttpClient struct {
	Requester
}

func (hr *HTTPRequester) NodeURL() string {
	return hr.url
}

func printBuffer(b []byte) string {
	st := "["
	for _, byt := range b {
		st += strconv.Itoa(int(byt)) + ","
	}
	st = st[:len(st)-1] + "]"
	return st
}

func (m HTTPRequester) AccountInfo(address string) (*accounts.AccountInfo, error) {
	str := fmt.Sprintf(`{ "address": "0x%s"}`, address)
	output, err := m.Get("/nonce", str, true)
	if err != nil {
		return nil, err
	}

	acc := accounts.AccountInfo{}
	if val, ok := output["value"]; ok {
		acc.Nonce = val.(string)
	} else {
		return nil, fmt.Errorf("cant get nonce %v", output)
	}

	output, err = m.Get("/balance", str, true)
	if err != nil {
		return nil, err
	}

	if val, ok := output["value"]; ok {
		acc.Balance = val.(string)
	} else {
		return nil, fmt.Errorf("cant get balance %v", output)
	}

	return &acc, nil
}

type NodeInfo struct {
	Synced                 bool
	SyncedLayer            string
	CurrentLayer           string
	VerifiedLayer          string
	Peers                  string
	MinPeers               string
	MaxPeers               string
	LibonomyDatadir        string
	LibonomyStatus         string
	LibonomyCoinbase       string
	LibonomyRemainingBytes string
}

func (m HTTPRequester) NodeInfo() (*NodeInfo, error) {
	nodeStatus, err := m.Get("/nodestatus", "", true)
	if err != nil {
		return nil, err
	}

	stats, err := m.Get("/stats", "", true)
	if err != nil {
		return nil, err
	}

	info := &NodeInfo{
		SyncedLayer:            "0",
		CurrentLayer:           "0",
		VerifiedLayer:          "0",
		Peers:                  "0",
		MinPeers:               "0",
		MaxPeers:               "0",
		LibonomyRemainingBytes: "0",
	}

	if val, ok := nodeStatus["synced"]; ok {
		info.Synced = val.(bool)
	}
	if val, ok := nodeStatus["syncedLayer"]; ok {
		info.SyncedLayer = val.(string)
	}
	if val, ok := nodeStatus["currentLayer"]; ok {
		info.CurrentLayer = val.(string)
	}
	if val, ok := nodeStatus["verifiedLayer"]; ok {
		info.VerifiedLayer = val.(string)
	}
	if val, ok := nodeStatus["peers"]; ok {
		info.Peers = val.(string)
	}
	if val, ok := nodeStatus["minPeers"]; ok {
		info.MinPeers = val.(string)
	}
	if val, ok := nodeStatus["maxPeers"]; ok {
		info.MaxPeers = val.(string)
	}
	if val, ok := stats["dataDir"]; ok {
		info.LibonomyDatadir = val.(string)
	}
	if val, ok := stats["status"]; ok {
		switch val.(float64) {
		case 1:
			info.LibonomyStatus = "`idle`"
		case 2:
			info.LibonomyStatus = "`in-progress`"
		case 3:
			info.LibonomyStatus = "`done`"
		}
	}
	if val, ok := stats["coinbase"]; ok {
		info.LibonomyCoinbase = val.(string)
	}
	if val, ok := stats["remainingBytes"]; ok {
		info.LibonomyRemainingBytes = val.(string)
	}

	return info, nil
}

func (m HTTPRequester) Send(b []byte) (string, error) {
	str := fmt.Sprintf(`{ "tx": %s}`, printBuffer(b))
	res, err := m.Get("/submittransaction", str, true)
	if err != nil {
		return "", err
	}

	val, ok := res["id"]
	if !ok {
		return "", errors.New("failed to submit tx")
	}
	return val.(string), nil
}

func (m HTTPRequester) Rebel(datadir string, space uint, coinbase string) error {
	str := fmt.Sprintf(`{ "logicalDrive": "%s", "commitmentSize": %d, "coinbase": "%s"}`, datadir, space, coinbase)
	_, err := m.Get("/startmining", str, true)
	if err != nil {
		return err
	}

	return nil
}

func (m HTTPRequester) ListTxs(address string) ([]string, error) {
	str := fmt.Sprintf(`{ "account": { "address": "%s"} }`, address)
	res, err := m.Get("/accounttxs", str, true)
	if err != nil {
		return nil, err
	}

	txs := make([]string, 0)
	val, ok := res["txs"]
	if !ok {
		return txs, nil
	}
	for _, val := range val.([]interface{}) {
		txs = append(txs, val.(string))
	}
	return txs, nil
}

func (m HTTPRequester) SetCoinbase(coinbase string) error {
	str := fmt.Sprintf(`{ "address": "%s"}`, coinbase)
	_, err := m.Get("/setawardsaddr", str, true)
	if err != nil {
		return err
	}

	return nil
}

func (m HTTPRequester) Sanity() error {
	_, err := m.Get("/example/echo", "", false)
	if err != nil {
		return err
	}

	return nil
}
