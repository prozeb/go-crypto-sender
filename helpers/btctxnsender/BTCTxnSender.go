package btctxnsender

import (
	"math/big"
	"strings"

	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/addressinfo"
	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/netchain"
	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/txutil"
	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/wallet"
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

	if err != nil {
		return "", err
	}
	return txID, nil
}

func PrivateKeyToAddress(privateKey string, isTestnet bool) (string, error) {

	net := netchain.MainNet
	if isTestnet {
		net = netchain.Signet
	}
	return wallet.AddressFromPrivateKey(privateKey, net)
}
func SatoshiToBigFloat(sat int64) *big.Int {
	f := new(big.Int).SetInt64(sat)
	divisor := big.NewInt(1e8)
	return new(big.Int).Quo(f, divisor)
}
func GetBalance(address string, isTestnet bool, apiKey string) (*big.Int, error) {

	net := netchain.MainNet
	if isTestnet {
		net = netchain.Signet
	}

	balance, err := addressinfo.GetBalance(address, net, apiKey)
	if err != nil {
		return nil, err
	}
	return big.NewInt(balance), nil
}
