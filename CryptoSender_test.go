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

	privatekeyWallet1 := "6500438c23646043765da05053bda643f771e3576e62165c3f1ec5000aeb17c1"
	// walletAddress1 := "0x6a89744e01be265b67c22a774947e751411d73de"
	walletAddress2 := "0x9cdc81ea100b15d24daf67f4c966fafbc8740abc"
	// privatekeyWallet2 := "50fe2d96a0f46faceceff948f490bea92ad9081ac921b5a940eca033b9f7a228"
	tokenAddress := "0x74872C11B0DA3E5090eA7E450874443Fa3729eD1"
	// wallet3 := "0xe9091bf58578be8cd342ac3bf462985ebfa103b3"
	tokenDecimals := 9
	networksAndRPCs := map[types.Network]string{
		types.BSC_TESTNET: "https://data-seed-prebsc-1-s1.binance.org:8545",
	}
	client, err := NewTxnMakerClient(networksAndRPCs)
	if err != nil {
		t.Fatal(err)
	}

	// transferTxn, err := client.TransferNative(context.TODO(), networks.NativeTxnOpts{
	// 	PrivateKey: privatekeyWallet2,
	// 	SendAll:    true,
	// 	To:         walletAddress1,
	// 	Network:    types.BSC_TESTNET,
	// })
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// fmt.Println("TransferNative hash", transferTxn)

	// ============================ Approve ===========================
	// approval, err := client.ApproveToken(context.Background(), networks.ApproveTokenOpts{
	// 	Network:         types.BSC_TESTNET,
	// 	ContractAddress: tokenAddress,
	// 	IsInfinite:      true,
	// 	Spender:         walletAddress2,
	// 	Allowance:       "1000",
	// 	PrivateKey:      privatekeyWallet1,
	// 	Decimals:        tokenDecimals,
	// })
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// fmt.Println("approvaltxn hash", approval)

	// // =========================== TransferFrom ===========================
	// transferFrom, err := client.TransferFrom(context.Background(), networks.TransferFromOpts{
	// 	PrivateKey:      privatekeyWallet2,
	// 	ContractAddress: tokenAddress,
	// 	Amount:          "5",
	// 	Decimals:        tokenDecimals,
	// 	Destination:     wallet3,
	// 	FromAddress:     walletAddress1,
	// 	Network:         types.BSC_TESTNET,
	// 	SendAll:         false,
	// })
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// fmt.Println("transferFrom hash", transferFrom)

	// =========================== Transfer ===========================
	transfer, err := client.BuildTransferTokenTxn(context.Background(), networks.TransferTokenOpts{
		Network:         types.BSC_TESTNET,
		ContractAddress: tokenAddress,
		To:              walletAddress2,
		Amount:          "1000",
		PrivateKey:      privatekeyWallet1,
		Decimals:        tokenDecimals,
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("transfertxn hash", transfer)
}

func TestBTCNativeBalanceTest(t *testing.T) {
	networksAndRPCs := map[types.Network]string{
		types.BTC_TESTNET: "",
	}
	client, err := NewTxnMakerClient(networksAndRPCs)
	if err != nil {
		t.Fatal(err)
	}

	balance, err := client.GetNativeBalance(context.Background(), networks.NativeBalanceOpts{
		Address: "tb1q2vqr5c9m400k0c99hd5f85j5x9jrdexuua8udj",
		Network: types.BTC_TESTNET,
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("balance", balance)
}

func TestBTCNativeTransferTest(t *testing.T) {
	networksAndRPCs := map[types.Network]string{
		types.BTC_TESTNET: "",
	}
	client, err := NewTxnMakerClient(networksAndRPCs)
	if err != nil {
		t.Fatal(err)
	}
	buildTxn, err := client.BuildTransferNativeTxn(context.Background(), networks.NativeTxnOpts{
		PrivateKey: "",
		To:         "",
		Network:    types.BTC_TESTNET,
		SendAll:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	txn, err := client.BroadcastTxn(context.Background(), buildTxn)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("txn", txn)
}

func TestGasNativeTransferTest(t *testing.T) {
	// networksAndRPCs := map[types.Network]string{
	// 	types.SEPOLIA: "https://rpc.ankr.com/eth_sepolia/3e16dce6de0d525f6232b9e7f3c462fdbc4bc3974c9829cb0ac91083000c5af1",
	// }
	// client, err := NewTxnMakerClient(networksAndRPCs)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// buildTxn, err := client.BuildTransferNativeTxn(context.Background(), networks.NativeTxnOpts{
	// 	PrivateKey: "e6c2fca2890cb0873053d584029f23478bbdf616764dbd87349a68cf0efa708c",
	// 	To:         "0x82f9745d366fedf8b3f2d5bffcdf1e73425dcf58",
	// 	Network:    types.SEPOLIA,
	// 	SendAll:    false,
	// 	Value:      "47623952376",
	// })
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// txn, err := client.BroadcastTxn(context.Background(), buildTxn)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// fmt.Println("txn", txn)
}
