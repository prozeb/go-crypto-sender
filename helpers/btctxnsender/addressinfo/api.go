package addressinfo

import "github.com/prozeb/go-crypto-sender/helpers/btctxnsender/netchain"

type Address struct {
	Balance int64  `json:"balance"`
	UTXOs   []UTXO `json:"utxos"`
}

type UTXO struct {
	TxID     string `json:"txid"`
	Pbscript string `json:"scriptpubkey"`
	Balance  int64  `json:"amount"`
	TxOutIdx int    `json:"vout"`
}

type Fetch func(address string, net netchain.Net, apiKey string) (Address, error)

// GetSatoshiPerByte returns minimum 'good-enough' satoshi per byte rate.
type GetSatoshiPerByte func(net netchain.Net) (int, error)
