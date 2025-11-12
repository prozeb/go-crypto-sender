package networks

import (
	"context"

	"github.com/prozeb/go-crypto-sender/types"
)

type NativeTxnOpts struct {
	PrivateKey string
	To         string
	Value      float64
	Network    types.Network
	SendAll    bool
}

type TransferTokenOpts struct {
	PrivateKey      string
	ContractAddress string
	Amount          string
	Decimals        int
	To              string
	Network         types.Network
	SendAll         bool
}

type ApproveTokenOpts struct {
	PrivateKey      string
	ContractAddress string
	Spender         string
	Allowance       string
	Decimals        int
	IsInfinite      bool
	Network         types.Network
}

type TransferFromOpts struct {
	PrivateKey      string
	ContractAddress string
	Amount          string
	Decimals        int
	Destination     string
	FromAddress     string
	Network         types.Network
	SendAll         bool
}

type CallTokenFunctionOpts struct {
	ContractAddress string
	FunctionName    string
	Network         types.Network
}

type AbstractClient interface {
	TransferNative(ctx context.Context, opts NativeTxnOpts) (string, error)
	TransferToken(ctx context.Context, opts TransferTokenOpts) (string, error)
	ApproveToken(ctx context.Context, opts ApproveTokenOpts) (string, error)
	TransferFrom(ctx context.Context, opts TransferFromOpts) (string, error)
	CallTokenFunction(ctx context.Context, opts CallTokenFunctionOpts, args ...interface{}) (string, error)
}
