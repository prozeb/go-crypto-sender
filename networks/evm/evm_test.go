package evm

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestGetBlock(t *testing.T) {
	client, err := NewEVMTxnMakerClient("https://1rpc.io/sepolia")
	if err != nil {
		t.Fatal(err)
	}

	totalGas, gasLimit, gasPrice, err := client.GetGasEstimation(context.Background(), common.HexToAddress("0xa5c943ad8779ee412fad3019d11dfb04a0913abe"), common.HexToAddress("0x82f9745d366fedf8b3f2d5bffcdf1e73425dcf58"), "", big.NewInt(0.01*1e18))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("gas", totalGas)
	fmt.Println("gasLimit", gasLimit)
	fmt.Println("gasPrice", gasPrice)
}
