package wallet

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/netchain"
)

// func AddressFromPrivateKey(privKey string, net netchain.Net) (string, error) {
// 	wif, err := btcutil.DecodeWIF(privKey)
// 	if err != nil {
// 		return "", fmt.Errorf("couldn't decode private key")
// 	}

// 	addr, err := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeUncompressed(), net.GetBtcdNetParams())
// 	if err != nil {
// 		return "", fmt.Errorf("couldn't extract address from private key")
// 	}
// 	return addr.EncodeAddress(), nil
// }

func AddressFromPrivateKey(privHex string, net netchain.Net) (string, error) {
	wif, err := btcutil.DecodeWIF(privHex)
	if err != nil {
		return "", fmt.Errorf("invalid hex: %w", err)
	}

	priv := wif.PrivKey
	pub := priv.PubKey()
	// P2WPKH (native SegWit)
	pubKeyHash := btcutil.Hash160(pub.SerializeCompressed())
	chainParams := chaincfg.MainNetParams
	if net == netchain.Signet {
		chainParams = chaincfg.SigNetParams
	}
	addr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chainParams)
	if err != nil {
		return "", fmt.Errorf("address error: %w", err)
	}
	return addr.EncodeAddress(), nil
}

func IsAddressValid(address string, net netchain.Net) bool {
	_, err := btcutil.DecodeAddress(address, net.GetBtcdNetParams())
	return err == nil
}
