package networks

import (
	"context"
	"errors"
	"strconv"

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

func (b *BTCTxnMakerClient) TransferNative(ctx context.Context, opts NativeTxnOpts) (string, error) {

	formattedAmount, err := utils.AmountToChainUnit(strconv.FormatFloat(opts.Value, 'f', -1, 64), "8")
	if err != nil {
		return "", err
	}
	finalAmount, _ := formattedAmount.Float64()
	return btcctxnsender.SendBTCTxn(opts.PrivateKey, opts.To, finalAmount, opts.SendAll, b.IsTestnet, b.ankrApiKey)
}

func (b *BTCTxnMakerClient) TransferToken(ctx context.Context, opts TransferTokenOpts) (string, error) {
	return "", errors.New("not implemented")
}
func (b *BTCTxnMakerClient) ApproveToken(ctx context.Context, opts ApproveTokenOpts) (string, error) {
	return "", errors.New("not implemented")
}
func (b *BTCTxnMakerClient) CallTokenFunction(ctx context.Context, opts CallTokenFunctionOpts, args ...interface{}) (string, error) {
	return "", errors.New("not implemented")
}

func (b *BTCTxnMakerClient) TransferFrom(ctx context.Context, opts TransferFromOpts) (string, error) {
	return "", errors.New("not implemented")
}
