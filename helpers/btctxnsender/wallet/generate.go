package wallet

import (
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/netchain"
)

func New(net netchain.Net) (privateKeyWif, bitcoinAddress string) {
	// errors shouldn't happen
	priv, err := btcec.NewPrivateKey()
	check(err)
	wif, err := btcutil.NewWIF(priv, net.GetBtcdNetParams(), false)
	check(err)
	addr, err := btcutil.NewAddressPubKey(priv.PubKey().SerializeUncompressed(), net.GetBtcdNetParams())
	check(err)
	return wif.String(), addr.EncodeAddress()
}

func NewAddress(net netchain.Net) string {
	_, address := New(net)
	return address
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
