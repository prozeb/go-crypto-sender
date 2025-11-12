package utils

import (
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/prozeb/go-crypto-sender/types"
)

func IsEVMNetwork(network types.Network) bool {
	for _, n := range types.EVMNetworks {
		if n == network {
			return true
		}
	}
	return false
}

// ✅ Convert Wei → Token (divide by 10^decimals)
func ChainUnitToAmount(valueStr string, tokenDecimal string) (float64, error) {
	valueInt := new(big.Int)
	if _, ok := valueInt.SetString(valueStr, 10); !ok {
		return 0, fmt.Errorf("invalid number string: %s", valueStr)
	}

	tokenDecimalInt, err := strconv.Atoi(tokenDecimal)
	if err != nil {
		return 0, fmt.Errorf("invalid token decimal: %s", tokenDecimal)
	}

	divisor := new(big.Float).SetFloat64(math.Pow10(tokenDecimalInt))
	valueFloat := new(big.Float).Quo(new(big.Float).SetInt(valueInt), divisor)

	result, _ := valueFloat.Float64()
	return result, nil
}

// ✅ Convert Token → Wei (multiply by 10^decimals)
func AmountToChainUnit(valueStr string, tokenDecimal string) (*big.Int, error) {
	valueFloat, ok := new(big.Float).SetString(valueStr)
	if !ok {
		return nil, fmt.Errorf("invalid number string: %s", valueStr)
	}

	tokenDecimalInt, err := strconv.Atoi(tokenDecimal)
	if err != nil {
		return nil, fmt.Errorf("invalid token decimal: %s", tokenDecimal)
	}

	multiplier := new(big.Float).SetFloat64(math.Pow10(tokenDecimalInt))
	valueWeiFloat := new(big.Float).Mul(valueFloat, multiplier)

	valueWei := new(big.Int)
	valueWeiFloat.Int(valueWei) // truncate fractional part
	return valueWei, nil
}
