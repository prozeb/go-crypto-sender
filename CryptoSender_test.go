package gocryptosender

import (
	"context"
	"fmt"
	"testing"

	"github.com/prozeb/go-crypto-sender/networks"
	"github.com/prozeb/go-crypto-sender/types"
)

func TestEVMTokenViewFunctionsTest(t *testing.T) {
	networksAndRPCs := map[types.Network]string{
		types.BSC_TESTNET: "https://data-seed-prebsc-1-s1.binance.org:8545",
	}
	client, err := NewTxnMakerClient(networksAndRPCs)
	if err != nil {
		t.Fatal(err)
	}

	tokenAddress := "0xd09e6c0779589e8f6104aedeec83b4053fb4ad2a"
	walletAddress := "0xB4a11CbD1cED349edaEb6042b90984b6223e7e0A"
	//=========================== Balance Check ===========================
	balance, err := client.CallTokenFunction(context.Background(), networks.CallTokenFunctionOpts{
		Network:         types.BSC_TESTNET,
		ContractAddress: tokenAddress,
		FunctionName:    "balanceOf"}, walletAddress)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("balance", balance)

	//=========================== Allowance Check ===========================
	allowance, err := client.CallTokenFunction(context.Background(), networks.CallTokenFunctionOpts{
		Network:         types.BSC_TESTNET,
		ContractAddress: tokenAddress,
		FunctionName:    "allowance"}, walletAddress, "0x1234567890123456789012345678901234567890")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("allowance", allowance)

	//=========================== Token Name ===========================
	tokenName, err := client.CallTokenFunction(context.Background(), networks.CallTokenFunctionOpts{
		Network:         types.BSC_TESTNET,
		ContractAddress: tokenAddress,
		FunctionName:    "name"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("tokenName", tokenName)

	//=========================== Token Symbol ===========================
	tokenSymbol, err := client.CallTokenFunction(context.Background(), networks.CallTokenFunctionOpts{
		Network:         types.BSC_TESTNET,
		ContractAddress: tokenAddress,
		FunctionName:    "symbol"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("tokenSymbol", tokenSymbol)

	//=========================== Token Decimals ===========================
	tokenDecimals, err := client.CallTokenFunction(context.Background(), networks.CallTokenFunctionOpts{
		Network:         types.BSC_TESTNET,
		ContractAddress: tokenAddress,
		FunctionName:    "decimals"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("tokenDecimals", tokenDecimals)

}

func TestTokenApproveFunctionsTest(t *testing.T) {

	privateKey := "1234567890123456789012345678901234567890123456789012345678901234"
	networksAndRPCs := map[types.Network]string{
		types.BSC_TESTNET: "https://data-seed-prebsc-1-s1.binance.org:8545",
	}
	client, err := NewTxnMakerClient(networksAndRPCs)
	if err != nil {
		t.Fatal(err)
	}

	tokenAddress := "0xd09e6c0779589e8f6104aedeec83b4053fb4ad2a"
	//=========================== Approve ===========================
	approval, err := client.ApproveToken(context.Background(), networks.ApproveTokenOpts{
		Network:         types.BSC_TESTNET,
		ContractAddress: tokenAddress,
		IsInfinite:      true,
		Spender:         "0x1234567890123456789012345678901234567890",
		Allowance:       "1000",
		PrivateKey:      privateKey,
		Decimals:        18,
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("approvaltxn hash", approval)

}
