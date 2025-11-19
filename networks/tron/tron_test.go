package tron

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/prozeb/go-crypto-sender/networks"
)

const privateKey1 = "28defdb6a876839a4a1c7786d2fd6e9c4447473c715eb0548234da022a89b84f"
const walletAddress1 = "TY3mjjBGbFmCAmWqHtwHk2hURKmyHi1L44"

const privateKey2 = "4fed6afa1640a70fe9d9c9538733e4bf41e1d491df744c3e615693f3fc1106ee"
const walletAddress2 = "TDCB5gVD2j8ydMzSvSKncgyU6d2vekrJTy"

const contractAddress = "TPZiV1hj4Mqwnwphksie6WmbxvBc3sQPvV"
const decimals = 6
const rpc = "https://api.shasta.trongrid.io"

func TestTransferTrx(t *testing.T) {

	client, err := NewTronTxnMakerClient(rpc)
	if err != nil {
		t.Fatal(err)
	}

	receiver := "TTE3T39c61MK8rN16YpABuwbsnLEwzWnz1"
	result, err := client.BuildTransferNativeTxn(context.Background(), networks.NativeTxnOpts{
		PrivateKey: privateKey1,
		To:         receiver,
		Value:      1,
	})
	if err != nil {
		t.Fatal(err)
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(jsonResult))
}

func TestTransferToken(t *testing.T) {

	client, err := NewTronTxnMakerClient(rpc)
	if err != nil {
		t.Fatal(err)
	}

	receiver := "TCE66ryFJDMAQmDT3s8XdTN2g9xTFoMw8s"
	decimals := 6

	result, err := client.BuildTransferTokenTxn(context.Background(), networks.TransferTokenOpts{
		PrivateKey:      privateKey1,
		ContractAddress: contractAddress,
		Amount:          "1",
		Decimals:        decimals,
		To:              receiver,
		SendAll:         true,
	})
	if err != nil {
		t.Fatal(err)
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("resultresult", string(jsonResult))

	resp, err := client.BroadcastTxn(context.Background(), result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}

func TestApproveFrom(t *testing.T) {

	client, err := NewTronTxnMakerClient(rpc)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.BuildApproveTokenTxn(context.Background(), networks.ApproveTokenOpts{
		PrivateKey:      privateKey1,
		ContractAddress: contractAddress,
		Spender:         walletAddress2,
		IsInfinite:      true,
		Decimals:        decimals,
	})

	if err != nil {
		t.Fatal(err)
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("result", string(jsonResult))

	resp, err := client.BroadcastTxn(context.Background(), result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}
