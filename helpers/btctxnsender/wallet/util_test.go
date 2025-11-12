package wallet

import (
	"testing"

	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/netchain"
	"github.com/stretchr/testify/assert"
)

func TestAddressFromPrivateKey(t *testing.T) {
	address, err := AddressFromPrivateKey("cSgmDBTpk543RCp331wJCpLfv63LFrxeb1ugo9zFh8etKw7fHFG5", netchain.Signet)
	assert.Nil(t, err)
	assert.EqualValues(t, "mgFv6afUVhrdd3D6mY2iyWzHVk5b64qTok", address)
}

func TestIsAddressValid(t *testing.T) {
	var valid = []string{
		"16ftSEQ4ctQFDtVZiUBusQUjRrGhM3JYwe",
		"3D2oetdNuZUqQHPJmcMDDHYoqkyNVsFk9r",
		"16rCmCmbuWDhPjWTrpQGaU3EPdZF7MTdUk",
		"3Cbq7aT1tY8kMxWLbitaG7yT6bPbKChq64",
		"3Nxwenay9Z8Lc9JBiywExpnEFiLp6Afp8v",
	}
	for _, a := range valid {
		assert.True(t, IsAddressValid(a, netchain.MainNet))
	}
	var invalid = []string{
		"",
		"a",
		"address",
		"1234567890",
		"mgFv6afUVhrdd3D6mY2iyWzHVk5b64qTok",
		"ghfteEc4gtQFDtVZiUBusQUjRrGhM3JYwe",
	}
	for _, a := range invalid {
		assert.False(t, IsAddressValid(a, netchain.MainNet))
	}
}
