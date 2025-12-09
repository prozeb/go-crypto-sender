package btc

import (
	"context"
	"errors"
	"math/big"
	"strconv"
	"strings"

	"github.com/prozeb/go-crypto-sender/networks"

	btcctxnsender "github.com/prozeb/go-crypto-sender/helpers/btctxnsender"
	"github.com/prozeb/go-crypto-sender/utils"
)

type BTCTxnMakerClient struct {
	IsTestnet  bool
	ankrApiKey string
}

func NewBTCTxnMakerClient(ankrApiKey string, isTestnet bool) (*BTCTxnMakerClient, error) {
	return &BTCTxnMakerClient{
		IsTestnet:  isTestnet,
		ankrApiKey: ankrApiKey,
	}, nil
}

func (b *BTCTxnMakerClient) BuildTransferNativeTxn(ctx context.Context, opts networks.NativeTxnOpts) (*networks.TxnBuildResult, error) {

	privateKeys := strings.Split(opts.PrivateKey, ",")

	if len(privateKeys) > 1 && !opts.SendAll {
		return nil, errors.New("multiple private keys are not supported for non-all transfer")
	}

	walletAddress, err := btcctxnsender.PrivateKeyToAddress(opts.PrivateKey, b.IsTestnet)
	if err != nil {
		return nil, err
	}
	walletBalance, err := b.GetNativeBalance(ctx, networks.NativeBalanceOpts{
		Address: walletAddress,
	})
	if err != nil {
		return nil, err
	}
	var formattedAmount *big.Int
	if opts.SendAll {
		formattedAmount = walletBalance
	} else {
		formattedAmount, err = utils.AmountToChainUnit(strconv.FormatFloat(opts.Value, 'f', -1, 64), "8")
		if err != nil {
			return nil, err
		}

		if walletBalance.Cmp(formattedAmount) < 0 {
			return nil, errors.New("not enough balance")
		}
	}
	finalAmount, _ := formattedAmount.Float64()

	data := ""
	if opts.SendAll {
		data = "all"
	}
	txnBuildResult := &networks.TxnBuildResult{
		Data:            data,
		To:              opts.To,
		Value:           big.NewInt(int64(finalAmount)),
		GasRequired:     big.NewInt(0),
		GasPrice:        big.NewInt(0),
		Network:         opts.Network,
		GasLimit:        0,
		IsSufficientGas: true,
		PrivateKey:      opts.PrivateKey,
	}
	return txnBuildResult, nil
}

func (b *BTCTxnMakerClient) BroadcastTxn(ctx context.Context, txn *networks.TxnBuildResult) (string, error) {
	amount, _ := txn.Value.Float64()
	return btcctxnsender.SendBTCTxn(txn.PrivateKey, txn.To, amount, txn.Data == "all", b.IsTestnet, b.ankrApiKey)
}
func (b *BTCTxnMakerClient) BuildTransferFromTxn(ctx context.Context, opts networks.TransferFromOpts) (*networks.TxnBuildResult, error) {
	return nil, errors.New("not implemented")
}
func (b *BTCTxnMakerClient) BuildTransferTokenTxn(ctx context.Context, opts networks.TransferTokenOpts) (*networks.TxnBuildResult, error) {
	return nil, errors.New("not implemented")
}
func (b *BTCTxnMakerClient) BuildApproveTokenTxn(ctx context.Context, opts networks.ApproveTokenOpts) (*networks.TxnBuildResult, error) {
	return nil, errors.New("not implemented")
}
func (b *BTCTxnMakerClient) CallTokenFunction(ctx context.Context, opts networks.CallTokenFunctionOpts, args ...interface{}) (string, error) {
	return "", errors.New("not implemented")
}

func (b *BTCTxnMakerClient) GetNativeBalance(ctx context.Context, opts networks.NativeBalanceOpts) (*big.Int, error) {
	return btcctxnsender.GetBalance(opts.Address, b.IsTestnet, b.ankrApiKey)
}
