package gocryptosender

import (
	"context"
	"fmt"

	"github.com/prozeb/go-crypto-sender/liberrors"
	"github.com/prozeb/go-crypto-sender/networks"
	"github.com/prozeb/go-crypto-sender/types"
	"github.com/prozeb/go-crypto-sender/utils"
)

type TxnMakerClient struct {
	NetworkClient map[types.Network]networks.AbstractClient
}

func NewTxnMakerClient(rpcs map[types.Network]string) (*TxnMakerClient, error) {
	t := &TxnMakerClient{}
	clients := make(map[types.Network]networks.AbstractClient)
	for network, rpc := range rpcs {
		if utils.IsEVMNetwork(network) {
			client, err := networks.NewEVMTxnMakerClient(rpc)
			if err != nil {
				return nil, err
			}
			clients[network] = client
		} else if network == types.TRON || network == types.SHASTA {
			client, err := networks.NewTronTxnMakerClient(rpc)
			if err != nil {
				return nil, err
			}
			clients[network] = client
		} else if network == types.BTC || network == types.BTC_TESTNET {
			client, err := networks.NewBTCTxnMakerClient(rpc, network == types.BTC_TESTNET)
			if err != nil {
				return nil, err
			}
			clients[network] = client
		}
	}
	t.NetworkClient = clients
	return t, nil
}

func (t *TxnMakerClient) MakeNativeTxn(ctx context.Context, opts networks.NativeTxnOpts) (string, error) {
	client, ok := t.NetworkClient[opts.Network]
	if !ok {
		fmt.Printf("network %s not supported", opts.Network)
		return "", liberrors.ErrUnsupported
	}
	return client.MakeNativeTxn(ctx, opts)
}

func (t *TxnMakerClient) TransferToken(ctx context.Context, opts networks.TransferTokenOpts) (string, error) {
	client, ok := t.NetworkClient[opts.Network]
	if !ok {
		fmt.Printf("network %s not supported", opts.Network)
		return "", liberrors.ErrUnsupported
	}
	return client.TransferToken(ctx, opts)
}

func (t *TxnMakerClient) ApproveToken(ctx context.Context, opts networks.ApproveTokenOpts) (string, error) {
	client, ok := t.NetworkClient[opts.Network]
	if !ok {
		fmt.Printf("network %s not supported", opts.Network)
		return "", liberrors.ErrUnsupported
	}
	return client.ApproveToken(ctx, opts)
}

func (t *TxnMakerClient) TransferFrom(ctx context.Context, opts networks.TransferFromOpts) (string, error) {
	client, ok := t.NetworkClient[opts.Network]
	if !ok {
		fmt.Printf("network %s not supported", opts.Network)
		return "", liberrors.ErrUnsupported
	}
	return client.TransferFrom(ctx, opts)
}

func (t *TxnMakerClient) CallTokenFunction(ctx context.Context, opts networks.CallTokenFunctionOpts, args ...interface{}) (string, error) {
	client, ok := t.NetworkClient[opts.Network]
	if !ok {
		fmt.Printf("network %s not supported", opts.Network)
		return "", liberrors.ErrUnsupported
	}
	return client.CallTokenFunction(ctx, opts, args...)
}
