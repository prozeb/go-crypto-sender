package addressinfo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/netchain"
)

func FetchFromAnkr(address string, net netchain.Net, ankrApiKey string) (Address, error) {

	balance, err := GetBalance(address, net, ankrApiKey)
	if err != nil {
		return Address{}, err
	}

	utxos, err := getUTXOs(address, net, ankrApiKey)
	if err != nil {
		return Address{}, err
	}

	resp := Address{
		Balance: balance,
		UTXOs:   utxos,
	}
	return resp, nil
}

func getUTXOs(address string, net netchain.Net, ankrApiKey string) ([]UTXO, error) {

	url := fmt.Sprintf("https://rpc.ankr.com/premium-http/%s/%s/api/v2/utxo/%s?confirmed=true", getAnkrNetworkKey(net), ankrApiKey, address)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	type UtxoResponse struct {
		Txid          string `json:"txid"`
		Vout          int    `json:"vout"`
		Value         string `json:"value"`
		Height        int    `json:"height"`
		Confirmations int    `json:"confirmations"`
		Coinbase      bool   `json:"coinbase"`
	}

	var info []UtxoResponse
	err = json.Unmarshal(body, &info)
	if err != nil {
		return nil, err
	}

	utxos := make([]UTXO, len(info))
	for i, utxo := range info {
		valueSats, err := strconv.ParseInt(utxo.Value, 10, 64)
		if err != nil {
			return nil, err
		}

		// derive PkScript from address
		chainParams := chaincfg.MainNetParams
		if net == netchain.TestNet {
			chainParams = chaincfg.SigNetParams
		}
		addr, err := btcutil.DecodeAddress(address, &chainParams)
		if err != nil {
			return nil, fmt.Errorf("invalid address: %w", err)
		}
		script, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, fmt.Errorf("script build failed: %w", err)
		}
		utxos[i] = UTXO{
			TxID:     utxo.Txid,
			Balance:  valueSats,
			Pbscript: fmt.Sprintf("%x", script),
			TxOutIdx: utxo.Vout,
		}
	}

	return utxos, nil
}

func GetBalance(address string, net netchain.Net, ankrApiKey string) (int64, error) {

	url := fmt.Sprintf("https://rpc.ankr.com/premium-http/%s/%s/api/v2/address/%s?details=basic", getAnkrNetworkKey(net), ankrApiKey, address)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	type BalanceResponse struct {
		Address            string `json:"address"`
		Balance            string `json:"balance"`
		TotalReceived      string `json:"totalReceived"`
		TotalSent          string `json:"totalSent"`
		UnconfirmedBalance string `json:"unconfirmedBalance"`
		UnconfirmedTxs     int    `json:"unconfirmedTxs"`
		Txs                int    `json:"txs"`
	}

	var balanceResp BalanceResponse
	err = json.Unmarshal(body, &balanceResp)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(balanceResp.Balance, 10, 64)
}

func getAnkrNetworkKey(net netchain.Net) string {
	networkKey := "btc_blockbook"

	if net == netchain.Signet {
		networkKey = "btc_blockbook_signet"
	}
	return networkKey
}

func BroadcastTxn(rawTx string, net netchain.Net, ankrApiKey string) (string, error) {
	url := fmt.Sprintf("https://rpc.ankr.com/premium-http/%s/%s/api/v2/sendtx/%s", getAnkrNetworkKey(net), ankrApiKey, rawTx)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type BroadcastResponse struct {
		Result string `json:"result"`
	}

	var broadcastResp BroadcastResponse
	err = json.Unmarshal(body, &broadcastResp)
	if err != nil {
		return "", err
	}

	return broadcastResp.Result, nil
}
