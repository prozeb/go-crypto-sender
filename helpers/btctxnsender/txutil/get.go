package txutil

import (
	"os"

	"github.com/blockcypher/gobcy/v2"
	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/netchain"
)

func GetConfirmations(txID string, net netchain.Net) (int, error) {
	btc := gobcy.API{Token: os.Getenv("BTC_API_KEY"), Coin: "btc", Chain: net.GetBlockcypherChain()}
	tx, err := btc.GetTX(txID, map[string]string{"limit": "1"})
	if err != nil {
		return 0, err
	}
	return tx.Confirmations, nil
}
