package gocryptosender

import (
	"context"
	"fmt"
	"math/big"

	"github.com/prozeb/go-crypto-sender/liberrors"
	"github.com/prozeb/go-crypto-sender/networks"

	"github.com/prozeb/go-crypto-sender/networks/btc"
	"github.com/prozeb/go-crypto-sender/networks/evm"
	"github.com/prozeb/go-crypto-sender/networks/tron"

	"github.com/prozeb/go-crypto-sender/types"
	"github.com/prozeb/go-crypto-sender/utils"
)

// TxnMakerClient provides a unified interface for creating and broadcasting blockchain transactions
// across multiple networks including EVM-compatible chains, TRON, and Bitcoin.
type TxnMakerClient struct {
	NetworkClient map[types.Network]networks.AbstractClient
}

// NewTxnMakerClient creates a new TxnMakerClient instance with the provided RPC URLs for different networks.
// It initializes the appropriate client for each network (EVM, TRON, or Bitcoin) based on the provided RPC URLs.
// Returns an error if any client initialization fails.
func NewTxnMakerClient(rpcs map[types.Network]string) (*TxnMakerClient, error) {
	t := &TxnMakerClient{}
	clients := make(map[types.Network]networks.AbstractClient)
	for network, rpc := range rpcs {
		if utils.IsEVMNetwork(network) {
			client, err := evm.NewEVMTxnMakerClient(rpc)
			if err != nil {
				return nil, err
			}
			clients[network] = client
		} else if network == types.TRON || network == types.SHASTA {
			client, err := tron.NewTronTxnMakerClient(rpc)
			if err != nil {
				return nil, err
			}
			clients[network] = client
		} else if network == types.BTC || network == types.BTC_TESTNET {
			client, err := btc.NewBTCTxnMakerClient(rpc, network == types.BTC_TESTNET)
			if err != nil {
				return nil, err
			}
			clients[network] = client
		}
	}
	t.NetworkClient = clients
	return t, nil
}

// BuildTransferNativeTxn constructs a transaction for transferring native currency (e.g., ETH, TRX) on the specified network.
// Returns the built transaction or an error if the network is not supported or if the operation fails.
func (t *TxnMakerClient) BuildTransferNativeTxn(ctx context.Context, opts networks.NativeTxnOpts) (*networks.TxnBuildResult, error) {
	client, ok := t.NetworkClient[opts.Network]
	if !ok {
		fmt.Printf("network %s not supported", opts.Network)
		return nil, liberrors.ErrUnsupported
	}
	return client.BuildTransferNativeTxn(ctx, opts)
}

// BuildTransferTokenTxn constructs a transaction for transferring ERC20/TRC20 tokens on the specified network.
// Returns the built transaction or an error if the network is not supported or if the operation fails.
func (t *TxnMakerClient) BuildTransferTokenTxn(ctx context.Context, opts networks.TransferTokenOpts) (*networks.TxnBuildResult, error) {
	client, ok := t.NetworkClient[opts.Network]
	if !ok {
		fmt.Printf("network %s not supported", opts.Network)
		return nil, liberrors.ErrUnsupported
	}
	return client.BuildTransferTokenTxn(ctx, opts)
}

// GetNativeBalance retrieves the native currency balance (e.g., ETH, TRX) of the specified address.
// Returns the balance as *big.Int or an error if the network is not supported or if the operation fails.
func (t *TxnMakerClient) GetNativeBalance(ctx context.Context, opts networks.NativeBalanceOpts) (*big.Int, error) {
	client, ok := t.NetworkClient[opts.Network]
	if !ok {
		fmt.Printf("network %s not supported", opts.Network)
		return nil, liberrors.ErrUnsupported
	}
	return client.GetNativeBalance(ctx, opts)
}

// BuildApproveTokenTxn creates an approval transaction for spending tokens on behalf of another address.
// This is typically used for DEX interactions or other smart contract interactions that require token spending approval.
// Returns the built approval transaction or an error if the network is not supported or if the operation fails.
func (t *TxnMakerClient) BuildApproveTokenTxn(ctx context.Context, opts networks.ApproveTokenOpts) (*networks.TxnBuildResult, error) {
	client, ok := t.NetworkClient[opts.Network]
	if !ok {
		fmt.Printf("network %s not supported", opts.Network)
		return nil, liberrors.ErrUnsupported
	}
	return client.BuildApproveTokenTxn(ctx, opts)
}

// BroadcastTxn sends a signed transaction to the network for processing.
// Returns the transaction hash if successful, or an error if the network is not supported or if broadcasting fails.
func (t *TxnMakerClient) BroadcastTxn(ctx context.Context, txn *networks.TxnBuildResult) (string, error) {
	client, ok := t.NetworkClient[txn.Network]
	if !ok {
		fmt.Printf("network %s not supported", txn.Network)
		return "", liberrors.ErrUnsupported
	}
	return client.BroadcastTxn(ctx, txn)
}

// BuildTransferFromTxn constructs a transaction to transfer tokens from one address to another on behalf of the token owner.
// This requires prior approval from the token owner.
// Returns the built transaction or an error if the network is not supported or if the operation fails.
func (t *TxnMakerClient) BuildTransferFromTxn(ctx context.Context, opts networks.TransferFromOpts) (*networks.TxnBuildResult, error) {
	client, ok := t.NetworkClient[opts.Network]
	if !ok {
		fmt.Printf("network %s not supported", opts.Network)
		return nil, liberrors.ErrUnsupported
	}
	return client.BuildTransferFromTxn(ctx, opts)
}

// CallTokenFunction executes a read-only function on a token contract and returns the result.
// This can be used to query token information like balanceOf, allowance, etc.
// Returns the function call result as a string or an error if the network is not supported or if the call fails.
func (t *TxnMakerClient) CallTokenFunction(ctx context.Context, opts networks.CallTokenFunctionOpts, args ...interface{}) (string, error) {
	client, ok := t.NetworkClient[opts.Network]
	if !ok {
		fmt.Printf("network %s not supported", opts.Network)
		return "", liberrors.ErrUnsupported
	}
	return client.CallTokenFunction(ctx, opts, args...)
}
