package txutil

import (
	"os"

	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/addressinfo"

	"github.com/blockcypher/gobcy/v2"
	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/netchain"
)

// Returns the hash of the broadcasted transaction.
func Broadcast(rawTx string, net netchain.Net) (string, error) {
	btc := gobcy.API{Token: os.Getenv("BTC_API_KEY"), Coin: "btc", Chain: net.GetBlockcypherChain()}
	tx, err := btc.PushTX(rawTx)
	if err != nil {
		return "", err
	}
	return tx.Trans.Hash, nil
}

func BroadcastAnkr(rawTx string, net netchain.Net, ankrApiKey string) (string, error) {
	return addressinfo.BroadcastTxn(rawTx, net, ankrApiKey)
}
