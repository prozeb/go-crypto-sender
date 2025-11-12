package main

import (
	"fmt"

	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/netchain"
	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/wallet"
)

func main() {
	privateKey, address := wallet.New(netchain.TestNet)
	fmt.Printf("Private Key: %s\nBitcoin address: %s\n", privateKey, address)
}
