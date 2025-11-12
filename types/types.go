package types

import (
	"crypto/ecdsa"
	"math/big"
)

type Wallet struct {
	PrivateKeyRaw string
	Address       string
	PrivateKey    *ecdsa.PrivateKey
	Nonce         uint64
	Balance       *big.Int
}
