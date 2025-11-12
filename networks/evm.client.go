package networks

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prozeb/go-crypto-sender/liberrors"
	localTypes "github.com/prozeb/go-crypto-sender/types"
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
	{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"}
]
`

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

func (c *EVMTxnMakerClient) MakeNativeTxn(ctx context.Context, opts NativeTxnOpts) (string, error) {
	client, err := c.getClient()
	if err != nil {
		return "", err
	}

	wallet, err := c.getWallet(ctx, opts.PrivateKey)
	if err != nil {
		return "", err
	}

	to := common.HexToAddress(opts.To)

	// ---------------- DYNAMIC GAS ESTIMATION ----------------

	// Estimate gas dynamically (usually ~21000 for simple transfers)
	totalGas, gasLimit, gasPrice, err := c.GetGasEstimation(ctx, common.HexToAddress(wallet.Address), to, "", nil)
	if err != nil {
		fmt.Println("error:getgas", err)
		return "", err
	}
	// --- Check if user can afford gas ---
	if wallet.Balance.Cmp(totalGas) < 0 {
		fmt.Println("error:balance", err)
		return "", liberrors.ErrInsufficientBalance
	}
	var value *big.Int
	if opts.SendAll {
		value = new(big.Int).Sub(wallet.Balance, totalGas)
		if value.Sign() <= 0 {
			return "", liberrors.ErrInsufficientBalance
		}
	} else {
		formattedAmount, err := utils.AmountToChainUnit(fmt.Sprintf("%.0f", opts.Value), "18")
		if err != nil {
			return "", err
		}
		totalCost := new(big.Int).Add(formattedAmount, totalGas)
		if wallet.Balance.Cmp(totalCost) < 0 {
			return "", liberrors.ErrInsufficientBalance
		}
		value = formattedAmount
	}

	tx := types.NewTransaction(wallet.Nonce, to, value, gasLimit, gasPrice, nil)

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

func (c *EVMTxnMakerClient) TransferToken(ctx context.Context, opts TransferTokenOpts) (string, error) {
	client, err := c.getClient()
	if err != nil {
		return "", err
	}

	wallet, err := c.getWallet(ctx, opts.PrivateKey)
	if err != nil {
		return "", err
	}
	fromAddr := crypto.PubkeyToAddress(wallet.PrivateKey.PublicKey)
	tokenBalance, err := c.CallTokenFunction(ctx, CallTokenFunctionOpts{
		ContractAddress: opts.ContractAddress,
		FunctionName:    "balanceOf",
		Network:         opts.Network,
	}, fromAddr)
	if err != nil {
		return "", err
	}
	amount := new(big.Int)

	if opts.SendAll {
		if tokenBalance == "0" {
			return "", liberrors.ErrInsufficientBalance
		}
		amount.SetString(tokenBalance, 10)
	} else {
		formattedAmount, err := utils.AmountToChainUnit(opts.Amount, strconv.Itoa(opts.Decimals))
		if err != nil {
			return "", err
		}
		tokenBalanceInWei := new(big.Int)
		tokenBalanceInWei.SetString(tokenBalance, 10)
		if tokenBalanceInWei.Cmp(formattedAmount) < 0 {
			return "", liberrors.ErrInsufficientBalance
		}
		amount.SetString(formattedAmount.String(), 10)
	}

	erc20ABI, _ := abi.JSON(bytes.NewReader([]byte(ERC20ABI)))
	to := common.HexToAddress(opts.To)
	data, err := erc20ABI.Pack("transfer", to, amount)
	if err != nil {
		fmt.Println("error:pack", err)
		return "", liberrors.ErrAbiError
	}

	contract := common.HexToAddress(opts.ContractAddress)

	totalGas, gasLimit, gasPrice, err := c.GetGasEstimation(ctx, fromAddr, contract, string(data), nil)
	if err != nil {
		fmt.Println("error:getgas", err)
		return "", err
	}
	if wallet.Balance.Cmp(totalGas) < 0 {
		fmt.Println("error:balance", err)
		return "", liberrors.ErrInsufficientBalance
	}
	tx := types.NewTransaction(wallet.Nonce, contract, big.NewInt(0), uint64(gasLimit), gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), wallet.PrivateKey)
	if err != nil {
		return "", err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	fmt.Printf("✅ Sent token transfer: %s\n", signedTx.Hash().Hex())
	return signedTx.Hash().Hex(), nil
}

func (c *EVMTxnMakerClient) ApproveToken(ctx context.Context, opts ApproveTokenOpts) (string, error) {
	client, err := c.getClient()
	if err != nil {
		return "", err
	}

	wallet, err := c.getWallet(ctx, opts.PrivateKey)
	if err != nil {
		return "", err
	}
	var amount *big.Int
	if opts.IsInfinite {
		amount = new(big.Int).Sub(
			new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil),
			big.NewInt(1),
		)
	} else {
		formattedAmount, err := utils.AmountToChainUnit(opts.Allowance, strconv.Itoa(opts.Decimals))
		if err != nil {
			return "", fmt.Errorf("failed to convert amount to wei: %w", err)
		}
		amount = formattedAmount
	}

	erc20ABI, _ := abi.JSON(bytes.NewReader([]byte(ERC20ABI)))
	to := common.HexToAddress(opts.Spender)
	data, err := erc20ABI.Pack("approve", to, amount)
	if err != nil {
		fmt.Println("error:pack", err)
		return "", liberrors.ErrAbiError
	}

	contract := common.HexToAddress(opts.ContractAddress)

	totalGas, gasLimit, gasPrice, err := c.GetGasEstimation(ctx, common.HexToAddress(wallet.Address), contract, string(data), nil)
	if err != nil {
		fmt.Println("error:getgas", err)
		return "", liberrors.ErrGasEstimation
	}

	if wallet.Balance.Cmp(totalGas) < 0 {
		fmt.Println("error:balance", err)
		return "", liberrors.ErrInsufficientBalance
	}
	tx := types.NewTransaction(wallet.Nonce, contract, big.NewInt(0), uint64(gasLimit), gasPrice, data)
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		fmt.Println("error:networkid", err)
		return "", liberrors.ErrFailedToGetNonce
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), wallet.PrivateKey)
	if err != nil {
		fmt.Println("error:signtx", err)
		return "", liberrors.ErrFailedToSignTx
	}

	if err := client.SendTransaction(ctx, signedTx); err != nil {
		fmt.Println("error:senttx", err)
		return "", liberrors.ErrFailedToSendTx
	}

	fmt.Printf("✅ Approve TX sent: %s\n", signedTx.Hash().Hex())
	return signedTx.Hash().Hex(), nil
}

func (c *EVMTxnMakerClient) TransferFrom(ctx context.Context, opts TransferFromOpts) (string, error) {

	client, err := c.getClient()
	if err != nil {
		return "", err
	}
	wallet, err := c.getWallet(ctx, opts.PrivateKey)
	if err != nil {
		return "", err
	}

	// --------------- STEP 1: Check allowance -----------------

	allowanceStr, err := c.CallTokenFunction(ctx, CallTokenFunctionOpts{
		ContractAddress: opts.ContractAddress,
		FunctionName:    "allowance",
		Network:         opts.Network,
	}, wallet.Address, opts.FromAddress)
	if err != nil {
		return "", err
	}
	allowance := new(big.Int)
	allowance.SetString(allowanceStr, 10)

	// --------------- STEP 2: Check balance -----------------

	balanceStr, err := c.CallTokenFunction(ctx, CallTokenFunctionOpts{
		ContractAddress: opts.ContractAddress,
		FunctionName:    "balanceOf",
		Network:         opts.Network,
	}, opts.FromAddress)
	if err != nil {
		fmt.Println("error:balance", err)
		return "", liberrors.ErrInsufficientBalance
	}
	tokenBalance := new(big.Int)
	tokenBalance.SetString(balanceStr, 10)

	// --------------- STEP 3: Determine transfer amount --------

	transferAmount := new(big.Int)

	if opts.SendAll {
		if tokenBalance.Cmp(allowance) < 0 {
			transferAmount.Set(tokenBalance)
		} else {
			transferAmount.Set(allowance)
		}
	} else {
		if opts.Amount == "" {
			return "", fmt.Errorf("amount is empty")
		}

		transferAmount.SetString(opts.Amount, 10)
	}

	if allowance.Cmp(transferAmount) < 0 {
		fmt.Println("error:allowance", err)
		return "", liberrors.ErrInsufficientAllowance
	}

	// Validate balance
	if tokenBalance.Cmp(transferAmount) < 0 {
		fmt.Println("error:balance", err)
		return "", liberrors.ErrInsufficientBalance
	}

	// Encode transferFrom call
	tokenABI, err := abi.JSON(strings.NewReader(ERC20ABI))
	if err != nil {
		fmt.Println("error:erc20abi", err)
		return "", liberrors.ErrAbiError
	}

	data, err := tokenABI.Pack("transferFrom", common.HexToAddress(opts.FromAddress), common.HexToAddress(opts.Destination), transferAmount)
	if err != nil {
		fmt.Println("error:pack", err)
		return "", liberrors.ErrAbiError
	}

	contractAddress := common.HexToAddress(opts.ContractAddress)

	totalGas, gasLimit, gasPrice, err := c.GetGasEstimation(ctx, common.HexToAddress(wallet.Address), contractAddress, string(data), nil)
	if err != nil {
		fmt.Println("error:gasprice", err)
		return "", liberrors.ErrGasEstimation
	}

	if wallet.Balance.Cmp(totalGas) < 0 {
		fmt.Println("error:balance", err)
		return "", liberrors.ErrInsufficientBalance
	}

	// --------------- STEP 5: Build and send tx ----------------
	tx := types.NewTransaction(wallet.Nonce, contractAddress, big.NewInt(0), gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		fmt.Println("error:networkid", err)
		return "", liberrors.ErrFailedToGetNonce
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

func (c *EVMTxnMakerClient) CallTokenFunction(ctx context.Context, opts CallTokenFunctionOpts, args ...interface{}) (string, error) {
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

func (c *EVMTxnMakerClient) getWallet(ctx context.Context, privateKey string) (wallet *localTypes.Wallet, err error) {
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
	return &localTypes.Wallet{
		Address:       address.String(),
		PrivateKeyRaw: privateKey,
		PrivateKey:    parsedPrivateKey,
		Nonce:         nonce,
		Balance:       balance,
	}, nil
}
