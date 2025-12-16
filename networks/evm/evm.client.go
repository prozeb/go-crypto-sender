package evm

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prozeb/go-crypto-sender/liberrors"
	"github.com/prozeb/go-crypto-sender/networks"
	"github.com/prozeb/go-crypto-sender/utils"
)

const ERC20ABI = `
[
	{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"type":"function"},
	{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"type":"function"},
	{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint256"}],"type":"function"},
	{"constant":true,"inputs":[{"name":"owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"type":"function"},
	{"constant":true,"inputs":[{"name":"owner","type":"address"},{"name":"spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"type":"function"},
	{"constant":false,"inputs":[{"name":"spender","type":"address"},{"name":"value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"type":"function"},
	{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"},
	{"constant":false,"inputs":[{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"type":"function"}
]
`

type Wallet struct {
	PrivateKeyRaw string
	Address       string
	PrivateKey    *ecdsa.PrivateKey
	Nonce         uint64
	Balance       *big.Int
}

type EVMTxnMakerClient struct {
	Rpc string
}

func NewEVMTxnMakerClient(rpc string) (*EVMTxnMakerClient, error) {
	client := &EVMTxnMakerClient{
		Rpc: rpc,
	}
	_, err := client.GetBlock(context.Background())
	if err != nil {
		fmt.Println("error:getblock", err)
		return nil, liberrors.ErrGetBlock
	}

	return client, nil
}

func (c *EVMTxnMakerClient) BuildTransferNativeTxn(ctx context.Context, opts networks.NativeTxnOpts) (*networks.TxnBuildResult, error) {

	wallet, err := c.getWallet(ctx, opts.PrivateKey)
	if err != nil {
		return nil, err
	}

	to := common.HexToAddress(opts.To)
	totalGas, gasLimit, gasPrice, err := c.GetGasEstimation(ctx, common.HexToAddress(wallet.Address), to, "", big.NewInt(.000000001*1e18))
	if err != nil {
		fmt.Println("error:getgas", err)
		return nil, liberrors.ErrGasEstimation
	}
	// --- Check if user can afford gas ---
	if wallet.Balance.Cmp(totalGas) < 0 {
		fmt.Println("error:balance", err)
		return nil, liberrors.ErrInsufficientBalance
	}
	var value *big.Int
	if opts.SendAll {
		value = new(big.Int).Sub(wallet.Balance, totalGas)
		if value.Sign() <= 0 {
			return nil, liberrors.ErrInsufficientBalance
		}
	} else {
		if opts.IsAmountInChainUnit {
			value = big.NewInt(int64(opts.Value))
		} else {
			formattedAmount, err := utils.AmountToChainUnit(fmt.Sprintf("%.0f", opts.Value), "18")
			if err != nil {
				return nil, err
			}
			value = formattedAmount
		}
		totalCost := new(big.Int).Add(value, totalGas)
		if wallet.Balance.Cmp(totalCost) < 0 {
			return nil, liberrors.ErrInsufficientBalance
		}
	}

	totalAmountToBeSpent := new(big.Int).Add(value, totalGas)

	txnBuildResult := &networks.TxnBuildResult{
		From:        wallet.Address,
		To:          opts.To,
		Value:       value,
		GasRequired: totalGas,
		GasPrice:    gasPrice,
		Network:     opts.Network,

		GasLimit:        gasLimit,
		IsSufficientGas: wallet.Balance.Cmp(totalAmountToBeSpent) >= 0,
		PrivateKey:      opts.PrivateKey,
	}
	return txnBuildResult, nil

}

func (c *EVMTxnMakerClient) BuildTransferTokenTxn(ctx context.Context, opts networks.TransferTokenOpts) (*networks.TxnBuildResult, error) {

	wallet, err := c.getWallet(ctx, opts.PrivateKey)
	if err != nil {
		return nil, err
	}
	fromAddr := crypto.PubkeyToAddress(wallet.PrivateKey.PublicKey)
	tokenBalance, err := c.CallTokenFunction(ctx, networks.CallTokenFunctionOpts{
		ContractAddress: opts.ContractAddress,
		FunctionName:    "balanceOf",
		Network:         opts.Network,
	}, fromAddr)
	if err != nil {
		return nil, err
	}
	amount := new(big.Int)

	if opts.SendAll {
		if tokenBalance == "0" {
			return nil, liberrors.ErrInsufficientBalance
		}
		amount.SetString(tokenBalance, 10)
	} else {
		if opts.IsAmountInChainUnit {
			amount.SetString(opts.Amount, 10)
		} else {
			formattedAmount, err := utils.AmountToChainUnit(opts.Amount, strconv.Itoa(opts.Decimals))
			if err != nil {
				return nil, err
			}
			amount = formattedAmount
		}
		tokenBalanceInWei := new(big.Int)
		tokenBalanceInWei.SetString(tokenBalance, 10)
		if tokenBalanceInWei.Cmp(amount) < 0 {
			return nil, liberrors.ErrInsufficientBalance
		}
	}

	erc20ABI, _ := abi.JSON(bytes.NewReader([]byte(ERC20ABI)))
	to := common.HexToAddress(opts.To)
	data, err := erc20ABI.Pack("transfer", to, amount)
	if err != nil {
		fmt.Println("error:pack", err)
		return nil, liberrors.ErrAbiError
	}

	contract := common.HexToAddress(opts.ContractAddress)

	totalGas, gasLimit, gasPrice, err := c.GetGasEstimation(ctx, fromAddr, contract, string(data), nil)
	if err != nil {
		fmt.Println("error:getgas", err)
		return nil, err
	}

	txnBuildResult := &networks.TxnBuildResult{
		Data:        string(data),
		From:        wallet.Address,
		To:          opts.ContractAddress,
		Value:       big.NewInt(0),
		GasRequired: totalGas,
		Network:     opts.Network,

		GasPrice:        gasPrice,
		GasLimit:        gasLimit,
		IsSufficientGas: wallet.Balance.Cmp(totalGas) >= 0,
		PrivateKey:      opts.PrivateKey,
	}
	return txnBuildResult, nil

}

func (c *EVMTxnMakerClient) BuildApproveTokenTxn(ctx context.Context, opts networks.ApproveTokenOpts) (*networks.TxnBuildResult, error) {

	wallet, err := c.getWallet(ctx, opts.PrivateKey)
	if err != nil {
		return nil, err
	}
	amount := big.NewInt(0)
	if opts.IsInfinite {
		amount = new(big.Int).Sub(
			new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil),
			big.NewInt(1),
		)
	} else {
		if opts.IsAmountInChainUnit {
			amount.SetString(opts.Allowance, 10)
		} else {
			formattedAmount, err := utils.AmountToChainUnit(opts.Allowance, strconv.Itoa(opts.Decimals))
			if err != nil {
				return nil, fmt.Errorf("failed to convert amount to wei: %w", err)
			}
			amount = formattedAmount
		}
	}

	erc20ABI, _ := abi.JSON(bytes.NewReader([]byte(ERC20ABI)))
	to := common.HexToAddress(opts.Spender)
	data, err := erc20ABI.Pack("approve", to, amount)
	if err != nil {
		fmt.Println("error:pack", err)
		return nil, liberrors.ErrAbiError
	}

	contract := common.HexToAddress(opts.ContractAddress)

	totalGas, gasLimit, gasPrice, err := c.GetGasEstimation(ctx, common.HexToAddress(wallet.Address), contract, string(data), nil)
	if err != nil {
		fmt.Println("error:getgas", err)
		return nil, liberrors.ErrGasEstimation
	}

	txnBuildResult := &networks.TxnBuildResult{
		Data:        string(data),
		From:        wallet.Address,
		To:          opts.ContractAddress,
		Value:       big.NewInt(0),
		GasRequired: totalGas,
		Network:     opts.Network,

		GasPrice:        gasPrice,
		GasLimit:        gasLimit,
		IsSufficientGas: wallet.Balance.Cmp(totalGas) >= 0,
		PrivateKey:      opts.PrivateKey,
	}
	return txnBuildResult, nil

}

func (c *EVMTxnMakerClient) BuildTransferFromTxn(ctx context.Context, opts networks.TransferFromOpts) (*networks.TxnBuildResult, error) {

	wallet, err := c.getWallet(ctx, opts.PrivateKey)
	if err != nil {
		return nil, err
	}

	// --------------- STEP 1: Check allowance -----------------

	allowanceStr, err := c.CallTokenFunction(ctx, networks.CallTokenFunctionOpts{
		ContractAddress: opts.ContractAddress,
		FunctionName:    "allowance",
		Network:         opts.Network,
	}, opts.FromAddress, wallet.Address)
	if err != nil {
		return nil, err
	}

	allowance := new(big.Int)
	allowance.SetString(allowanceStr, 10)

	// --------------- STEP 2: Check balance -----------------

	balanceStr, err := c.CallTokenFunction(ctx, networks.CallTokenFunctionOpts{
		ContractAddress: opts.ContractAddress,
		FunctionName:    "balanceOf",
		Network:         opts.Network,
	}, opts.FromAddress)
	if err != nil {
		fmt.Println("error:balance", err)
		return nil, liberrors.ErrInsufficientBalance
	}
	tokenBalance := new(big.Int)
	tokenBalance.SetString(balanceStr, 10)

	// --------------- STEP 3: Determine transfer amount --------
	//todo implement ErrInsufficientTokenBalance
	transferAmount := new(big.Int)
	if opts.SendAll {
		if tokenBalance.Cmp(allowance) < 0 {
			transferAmount.Set(tokenBalance)
		} else {
			transferAmount.Set(allowance)
		}
	} else {
		if opts.Amount == "" {
			return nil, fmt.Errorf("amount is empty")
		}
		if opts.IsAmountInChainUnit {
			transferAmount.SetString(opts.Amount, 10)
		} else {
			amount, err := utils.AmountToChainUnit(opts.Amount, strconv.Itoa(opts.Decimals))
			if err != nil {
				return nil, err
			}
			transferAmount.Set(amount)
		}
	}

	if allowance.Cmp(transferAmount) < 0 {
		fmt.Println("error:allowance", err)
		return nil, liberrors.ErrInsufficientAllowance
	}

	// Validate balance
	if tokenBalance.Cmp(transferAmount) < 0 {
		fmt.Println("error:balance", err)
		return nil, liberrors.ErrInsufficientBalance
	}

	// Encode transferFrom call
	tokenABI, err := abi.JSON(strings.NewReader(ERC20ABI))
	if err != nil {
		fmt.Println("error:erc20abi", err)
		return nil, liberrors.ErrAbiError
	}

	data, err := tokenABI.Pack("transferFrom", common.HexToAddress(opts.FromAddress), common.HexToAddress(opts.Destination), transferAmount)
	if err != nil {
		fmt.Println("error:pack", err)
		return nil, liberrors.ErrAbiError
	}

	contractAddress := common.HexToAddress(opts.ContractAddress)

	totalGas, gasLimit, gasPrice, err := c.GetGasEstimation(ctx, common.HexToAddress(wallet.Address), contractAddress, string(data), nil)
	if err != nil {
		fmt.Println("error:gasprice", err)
		return nil, liberrors.ErrGasEstimation
	}

	txnBuildResult := &networks.TxnBuildResult{
		Data:            string(data),
		From:            wallet.Address,
		To:              opts.ContractAddress,
		Value:           big.NewInt(0),
		Network:         opts.Network,
		GasRequired:     totalGas,
		GasPrice:        gasPrice,
		GasLimit:        gasLimit,
		IsSufficientGas: wallet.Balance.Cmp(totalGas) >= 0,
		PrivateKey:      opts.PrivateKey,
	}
	return txnBuildResult, nil

}

func (c *EVMTxnMakerClient) GetNativeBalance(ctx context.Context, opts networks.NativeBalanceOpts) (*big.Int, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}
	return client.BalanceAt(ctx, common.HexToAddress(opts.Address), nil)
}

func (c *EVMTxnMakerClient) GetGasEstimation(ctx context.Context,
	fromAddress common.Address,
	toAddress common.Address, data string, value *big.Int) (totalGas *big.Int, gasLimit uint64, gasPrice *big.Int, err error) {
	client, err := c.getClient()
	if err != nil {
		return nil, 0, nil, err
	}
	if data == "" {
		data = "0x"
	}
	gasPrice, err = client.SuggestGasPrice(ctx)

	if err != nil {
		fmt.Println("error:gasprice", err)
		return nil, 0, nil, liberrors.ErrGasEstimation
	}
	msg := ethereum.CallMsg{
		From:  fromAddress,
		To:    &toAddress,
		Data:  []byte(data),
		Value: value,
	}

	gasLimit, err = client.EstimateGas(ctx, msg)
	if err != nil {
		fmt.Println("error:gaslimit", err)

		return nil, 0, nil, liberrors.ErrGasEstimation
	}

	totalGas = new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gasLimit))

	return totalGas, gasLimit, gasPrice, nil
}

func (c *EVMTxnMakerClient) CallTokenFunction(ctx context.Context, opts networks.CallTokenFunctionOpts, args ...interface{}) (string, error) {
	client, err := c.getClient()
	if err != nil {
		return "", err
	}

	erc20ABI, err := abi.JSON(bytes.NewReader([]byte(ERC20ABI)))
	if err != nil {
		fmt.Println("error:erc20abi", err)
		return "", liberrors.ErrAbiError
	}

	// Convert string addresses to common.Address
	for i, arg := range args {
		if addr, ok := arg.(string); ok {
			if common.IsHexAddress(addr) {
				args[i] = common.HexToAddress(addr)
			}
		}
	}

	data, err := erc20ABI.Pack(opts.FunctionName, args...)
	if err != nil {
		fmt.Println("error:Pack", err)
		return "", liberrors.ErrAbiError
	}

	toAddr := common.HexToAddress(opts.ContractAddress)
	callMsg := ethereum.CallMsg{To: &toAddr, Data: data}
	res, err := client.CallContract(ctx, callMsg, nil)
	if err != nil {
		fmt.Println("error:CallContract", err)
		return "", liberrors.ErrAbiError
	}

	// Decode the ABI-encoded result
	outputs, err := erc20ABI.Unpack(opts.FunctionName, res)
	if err != nil {
		fmt.Println("error:Unpack", err)
		return "", liberrors.ErrAbiError
	}

	if len(outputs) == 0 {
		fmt.Println("error:Unpack", err)
		return "", liberrors.ErrAbiError
	}

	// Handle different return types
	switch v := outputs[0].(type) {
	case *big.Int:
		return v.String(), nil
	case common.Address:
		return v.Hex(), nil
	case string:
		return v, nil
	default:
		fmt.Println("error:default", v)
		return "", liberrors.ErrAbiError
	}
}

func (c *EVMTxnMakerClient) BroadcastTxn(ctx context.Context, txn *networks.TxnBuildResult) (string, error) {
	wallet, err := c.getWallet(ctx, txn.PrivateKey)
	if err != nil {
		return "", err
	}

	totalValue := big.NewInt(0).Add(txn.GasRequired, txn.Value)
	if wallet.Balance.Cmp(totalValue) < 0 {
		fmt.Println("error:balance", err)
		return "", liberrors.ErrInsufficientBalance
	}
	tx := types.NewTransaction(wallet.Nonce, common.HexToAddress(txn.To), txn.Value, txn.GasLimit, txn.GasPrice, []byte(txn.Data))

	client, err := c.getClient()
	if err != nil {
		return "", err
	}
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		fmt.Println("error:networkid", err)
		return "", err
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), wallet.PrivateKey)
	if err != nil {
		fmt.Println("error:signtx", err)
		return "", liberrors.ErrFailedToSignTx
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		fmt.Println("error:senttx", err)
		return "", liberrors.ErrFailedToSendTx
	}

	return signedTx.Hash().Hex(), nil
}

func (c *EVMTxnMakerClient) GetBlock(ctx context.Context) (*types.Block, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, liberrors.ErrGetBlock
	}
	return client.BlockByNumber(ctx, nil)
}

func (c *EVMTxnMakerClient) getClient() (*ethclient.Client, error) {
	client, err := ethclient.Dial(c.Rpc)
	if err != nil {
		return nil, liberrors.ErrRPCClient
	}

	return client, nil
}

func (c *EVMTxnMakerClient) getWallet(ctx context.Context, privateKey string) (wallet *Wallet, err error) {
	parsedPrivateKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, liberrors.ErrInvalidPrivateKey
	}

	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	address := crypto.PubkeyToAddress(parsedPrivateKey.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, address)
	if err != nil {
		return nil, liberrors.ErrFailedToGetNonce
	}

	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, liberrors.ErrFailedToGetBalance
	}
	return &Wallet{
		Address:       address.String(),
		PrivateKeyRaw: privateKey,
		PrivateKey:    parsedPrivateKey,
		Nonce:         nonce,
		Balance:       balance,
	}, nil
}
