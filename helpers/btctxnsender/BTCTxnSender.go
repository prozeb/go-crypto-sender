package btctxnsender

import (
	"fmt"
	"strings"

	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/netchain"
	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/txutil"
)

func SendBTCTxn(privateKey string, toAddress string, amount float64, sendAll bool, isTestnet bool, apiKey string) (string, error) {

	net := netchain.MainNet
	if isTestnet {
		net = netchain.Signet
	}
	privateKeys := strings.Split(privateKey, ",")

	rawTx, err := txutil.Create(txutil.CreateParams{
		PrivateKeys: privateKeys,
		Destination: toAddress,
		Amount:      int64(amount),
		Net:         net,
		SendAll:     sendAll,
		// MinerFee:    5000,
		ApiKey: apiKey,
	})

	if err != nil {
		return "", err
	}

	txID, err := txutil.BroadcastAnkr(rawTx, net, apiKey)
	fmt.Println("rawTxrawTxrawTx222", txID, err)

	if err != nil {
		return "", err
	}
	return txID, nil
}
