package networks

import (
	"context"
	"math/big"

	"github.com/prozeb/go-crypto-sender/types"
)

type NativeTxnOpts struct {
	PrivateKey          string
	To                  string
	Value               float64
	Network             types.Network
	SendAll             bool
	IsAmountInChainUnit bool
}

type TransferTokenOpts struct {
	PrivateKey          string
	ContractAddress     string
	Amount              string
	Decimals            int
	To                  string
	Network             types.Network
	SendAll             bool
	IsAmountInChainUnit bool
}

type ApproveTokenOpts struct {
	PrivateKey          string
	ContractAddress     string
	Spender             string
	Allowance           string
	Decimals            int
	IsInfinite          bool
	Network             types.Network
	IsAmountInChainUnit bool
}

type TransferFromOpts struct {
	PrivateKey          string
	ContractAddress     string
	Amount              string
	Decimals            int
	Destination         string
	FromAddress         string
	Network             types.Network
	SendAll             bool
	IsAmountInChainUnit bool
}

type NativeBalanceOpts struct {
	Network types.Network
	Address string
}

type CallTokenFunctionOpts struct {
	ContractAddress string
	FunctionName    string
	Network         types.Network
}

type TxnBuildResult struct {
	Data            string   `json:"data"`
	From            string   `json:"from"`
	To              string   `json:"to"`
	Value           *big.Int `json:"value"`
	GasRequired     *big.Int `json:"gas_required"`
	GasLimit        uint64   `json:"gas_limit"`
	GasPrice        *big.Int `json:"gas_price"`
	PrivateKey      string   `json:"private_key"`
	IsSufficientGas bool     `json:"is_sufficient_gas"`
	Network         types.Network
}

type AbstractClient interface {
	BuildTransferNativeTxn(ctx context.Context, opts NativeTxnOpts) (*TxnBuildResult, error)
	BuildTransferTokenTxn(ctx context.Context, opts TransferTokenOpts) (*TxnBuildResult, error)
	BuildApproveTokenTxn(ctx context.Context, opts ApproveTokenOpts) (*TxnBuildResult, error)
	BuildTransferFromTxn(ctx context.Context, opts TransferFromOpts) (*TxnBuildResult, error)
	GetNativeBalance(ctx context.Context, opts NativeBalanceOpts) (*big.Int, error)
	CallTokenFunction(ctx context.Context, opts CallTokenFunctionOpts, args ...interface{}) (string, error)
	BroadcastTxn(ctx context.Context, txn *TxnBuildResult) (string, error)
}

type AbstractTokenFunction interface {
	GasEstimate() (*big.Int, error)
	Send() (string, error)
}
