package tron

import (
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/sha3"

	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// pad 32 bytes left
func pad32Left(input []byte) []byte {
	padded := make([]byte, 32)
	copy(padded[32-len(input):], input)
	return padded
}

// keccak256 function selector (first 4 bytes)
func FunctionSelector(signature string) string {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(signature))
	return hex.EncodeToString(hash.Sum(nil)[:4])
}

// encodes a single Solidity value
func EncodeValue(solType string, value interface{}) (string, error) {
	switch solType {

	case "address":
		v := value.(string)
		v = strings.TrimPrefix(v, "0x")
		v = strings.TrimPrefix(v, "41") // Tron hex address starts with 41
		addrBytes, _ := hex.DecodeString(v)
		return hex.EncodeToString(pad32Left(addrBytes)), nil

	case "uint256":
		bi := new(big.Int)
		switch val := value.(type) {
		case int64:
			bi = big.NewInt(val)
		case uint64:
			bi = new(big.Int).SetUint64(val)
		case string:
			bi.SetString(val, 10)
		case *big.Int:
			bi = val
		}
		return hex.EncodeToString(pad32Left(bi.Bytes())), nil

	default:
		return "", fmt.Errorf("unsupported type: %s", solType)
	}
}

// EncodeABI — encode function selector + all args
func EncodeABI(signature string, types []string, args []interface{}) (string, error) {
	if len(types) != len(args) {
		return "", fmt.Errorf("parameter count mismatch")
	}

	selector := FunctionSelector(signature)
	encoded := ""

	for i, t := range types {
		v, err := EncodeValue(t, args[i])
		if err != nil {
			return "", err
		}
		encoded += v
	}

	return selector + encoded, nil
}

func TronBase58ToHex(addr string) string {
	decoded := base58.Decode(addr)
	hexAddr := hex.EncodeToString(decoded[:len(decoded)-4])
	return hexAddr
}
