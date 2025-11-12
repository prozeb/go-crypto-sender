package addressinfo

import (
	"testing"

	"github.com/prozeb/go-crypto-sender/helpers/btctxnsender/netchain"
	"github.com/stretchr/testify/assert"
)

func TestFetchFromBlockchain(t *testing.T) {
	t.Run(netchain.MainNet.String(), func(t *testing.T) {
		got, err := FetchFromBlockchain("3LQUu4v9z6KNch71j7kbj8GPeAGUo1FW6a", netchain.MainNet)
		assert.Nil(t, err)
		assert.Positive(t, got.Balance)
		assert.Positive(t, len(got.UTXOs))
	})
}

func TestGetSatoshiPerByteFromBlockchain(t *testing.T) {
	spb, err := GetSatoshiPerByteFromBlockchain(netchain.MainNet)
	if err != nil {
		t.Fatal(err)
	}
	assert.Positive(t, spb)
}
