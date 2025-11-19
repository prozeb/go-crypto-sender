package tron

import (
	"github.com/0x10f/go-tron/abi"
	"github.com/prozeb/go-crypto-sender/liberrors"
	"github.com/prozeb/go-crypto-sender/networks"

	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/prozeb/go-crypto-sender/utils"

	tronWallet "github.com/0x10f/go-tron/account"
	"github.com/0x10f/go-tron/address"
	tronClient "github.com/0x10f/go-tron/client"
	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/crypto"
)

// --------------------------- Types / Client ---------------------------
var erc20ABI abi.ABI = abi.ABI{
	Functions: map[string]abi.Function{
		// name() public view returns (string)
		"name": {
			Name:       "name",
			Mutability: "view",
			Inputs:     []abi.Value{},
			Outputs: []abi.Value{
				{Name: "", Type: "bytes32"}, // using bytes32 instead of string for compatibility
			},
		},
		// symbol() public view returns (string)
		"symbol": {
			Name:       "symbol",
			Mutability: "view",
			Inputs:     []abi.Value{},
			Outputs: []abi.Value{
				{Name: "", Type: "bytes32"},
			},
		},
		// decimals() public view returns (uint8)
		"decimals": {
			Name:       "decimals",
			Mutability: "view",
			Inputs:     []abi.Value{},
			Outputs: []abi.Value{
				{Name: "", Type: "uint256"},
			},
		},
		// totalSupply() public view returns (uint256)
		"totalSupply": {
			Name:       "totalSupply",
			Mutability: "view",
			Inputs:     []abi.Value{},
			Outputs: []abi.Value{
				{Name: "totalSupply", Type: "uint256"},
			},
		},
		// balanceOf(address account) public view returns (uint256)
		"balanceOf": {
			Name:       "balanceOf",
			Mutability: "view",
			Inputs: []abi.Value{
				{Name: "account", Type: "address"},
			},
			Outputs: []abi.Value{
				{Name: "balance", Type: "uint256"},
			},
		},
		// transfer(address to, uint256 value) public returns (bool)
		"transfer": {
			Name:       "transfer",
			Mutability: "nonpayable",
			Inputs: []abi.Value{
				{Name: "to", Type: "address"},
				{Name: "value", Type: "uint256"},
			},
			Outputs: []abi.Value{
				{Name: "success", Type: "bool"},
			},
		},
		// approve(address spender, uint256 value) public returns (bool)
		"approve": {
			Name:       "approve",
			Mutability: "nonpayable",
			Inputs: []abi.Value{
				{Name: "spender", Type: "address"},
				{Name: "value", Type: "uint256"},
			},
			Outputs: []abi.Value{
				{Name: "success", Type: "bool"},
			},
		},
		// allowance(address owner, address spender) public view returns (uint256)
		"allowance": {
			Name:       "allowance",
			Mutability: "view",
			Inputs: []abi.Value{
				{Name: "owner", Type: "address"},
				{Name: "spender", Type: "address"},
			},
			Outputs: []abi.Value{
				{Name: "remaining", Type: "uint256"},
			},
		},
		// transferFrom(address from, address to, uint256 value) public returns (bool)
		"transferFrom": {
			Name:       "transferFrom",
			Mutability: "nonpayable",
			Inputs: []abi.Value{
				{Name: "from", Type: "address"},
				{Name: "to", Type: "address"},
				{Name: "value", Type: "uint256"},
			},
			Outputs: []abi.Value{
				{Name: "success", Type: "bool"},
			},
		},
	},

	Events: map[string]abi.Event{
		"Transfer": {
			Name: "Transfer",
			Inputs: []abi.Value{
				{Name: "from", Type: "address", Indexed: true},
				{Name: "to", Type: "address", Indexed: true},
				{Name: "value", Type: "uint256"},
			},
		},
		"Approval": {
			Name: "Approval",
			Inputs: []abi.Value{
				{Name: "owner", Type: "address", Indexed: true},
				{Name: "spender", Type: "address", Indexed: true},
				{Name: "value", Type: "uint256"},
			},
		},
	},
}

type TokenTxnData struct {
	FunctionName string     `json:"functionName"`
	TokenAddress string     `json:"tokenAddress"`
	Arguments    []DataArgs `json:"arguments"`
}

type DataArgs struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type TronTxnMakerClient struct {
	Rpc        string
	httpClient *http.Client
	trxClient  *tronClient.Client
}

type TronWallet struct {
	address string
	account *tronWallet.LocalAccount
	Nonce   uint64
	balance *big.Int
}

func NewTronTxnMakerClient(rpc string) (*TronTxnMakerClient, error) {
	trxClient := tronClient.New(rpc)
	return &TronTxnMakerClient{
		Rpc:       rpc,
		trxClient: trxClient,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}, nil
}

// --------------------------- Public: Native TRX ---------------------------

// TransferNative creates, signs, and broadcasts a native TRX transfer.
func (c *TronTxnMakerClient) BuildTransferNativeTxn(ctx context.Context, opts networks.NativeTxnOpts) (*networks.TxnBuildResult, error) {

	wallet, err := c.getWallet(ctx, opts.PrivateKey)
	if err != nil {
		return nil, err
	}

	finalAmount, err := utils.AmountToChainUnit(fmt.Sprintf("%f", opts.Value), "8")
	if err != nil {
		return nil, err
	}
	gasAmount, err := utils.AmountToChainUnit(fmt.Sprintf("%f", 1.1*1e8), "8")
	if err != nil {
		return nil, err
	}

	if opts.SendAll {
		finalAmount = new(big.Int).Sub(wallet.balance, gasAmount)
		if finalAmount.Sign() <= 0 {
			return nil, liberrors.ErrInsufficientBalance
		}
	}
	totalAmountToBeSpent := new(big.Int).Add(finalAmount, gasAmount)

	result := &networks.TxnBuildResult{
		From:    wallet.address,
		To:      opts.To,
		Value:   finalAmount,
		Network: opts.Network,

		IsSufficientGas: wallet.balance.Cmp(totalAmountToBeSpent) >= 0,
		PrivateKey:      opts.PrivateKey,
	}

	return result, nil
}

// --------------------------- Public: TRC20 - Transfer ---------------------------

func (c *TronTxnMakerClient) BuildTransferTokenTxn(ctx context.Context, opts networks.TransferTokenOpts) (*networks.TxnBuildResult, error) {

	wallet, err := c.getWallet(ctx, opts.PrivateKey)
	if err != nil {
		return nil, err
	}

	tokenBalance, err := c.CallTokenFunction(ctx, networks.CallTokenFunctionOpts{
		ContractAddress: opts.ContractAddress,
		FunctionName:    "balanceOf",
	}, wallet.address)

	if err != nil {
		return nil, err
	}

	tokenBalanceInBigInt := new(big.Int)
	tokenBalanceInBigInt.SetString(tokenBalance, 10)

	finalAmount := new(big.Int)
	if opts.SendAll {
		finalAmount = tokenBalanceInBigInt
	} else {
		amount, err := utils.AmountToChainUnit(opts.Amount, fmt.Sprintf("%d", opts.Decimals))
		if err != nil {
			return nil, err
		}
		finalAmount = amount
	}

	if tokenBalanceInBigInt.Cmp(finalAmount) < 0 {
		return nil, liberrors.ErrInsufficientTokenBalance
	}

	selector := "transfer(address,uint256)"

	params, _ := EncodeABI(
		selector,
		[]string{"address", "uint256"},
		[]interface{}{
			"0x41" + TronBase58ToHex(opts.To), // hex address without base58
			finalAmount,                       // amount in token decimals
		},
	)

	feeInTrx, err := c.EstimateFeeDynamic(
		wallet.address,
		"",
		0,
		opts.ContractAddress,
		selector,
		params[8:], // only parameters (skip selector)
	)
	if err != nil {
		return nil, err
	}

	dataArgs := []DataArgs{
		{
			Value: opts.To,
			Type:  "address",
		},
		{
			Value: finalAmount.String(),
			Type:  "bigInt",
		},
	}
	input := TokenTxnData{
		FunctionName: "transfer",
		TokenAddress: opts.ContractAddress,
		Arguments:    dataArgs,
	}

	inputInStr, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	result := &networks.TxnBuildResult{
		Data:            string(inputInStr),
		From:            wallet.address,
		To:              opts.ContractAddress,
		Value:           big.NewInt(0),
		GasRequired:     feeInTrx,
		IsSufficientGas: wallet.balance.Cmp(feeInTrx) >= 0,
		Network:         opts.Network,

		PrivateKey: opts.PrivateKey,
	}
	return result, nil

}

// --------------------------- Public: TRC20 - Approve ---------------------------
func (c *TronTxnMakerClient) BuildApproveTokenTxn(ctx context.Context, opts networks.ApproveTokenOpts) (*networks.TxnBuildResult, error) {

	wallet, err := c.getWallet(ctx, opts.PrivateKey)
	if err != nil {
		return nil, err
	}

	amount := new(big.Int)
	if opts.IsInfinite {
		amount.Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))
	} else {
		finalAmount, err := utils.AmountToChainUnit(opts.Allowance, fmt.Sprintf("%d", opts.Decimals))
		if err != nil {
			return nil, err
		}
		amount = finalAmount
	}
	selector := "approve(address,uint256)"

	params, _ := EncodeABI(
		selector,
		[]string{"address", "uint256"},
		[]interface{}{
			"0x41" + TronBase58ToHex(opts.Spender), // hex address without base58
			amount,                                 // amount in token decimals
		},
	)

	feeInTrx, err := c.EstimateFeeDynamic(
		wallet.address,
		"",
		0,
		opts.ContractAddress,
		selector,
		params[8:], // only parameters (skip selector)
	)
	if err != nil {
		return nil, err
	}

	dataArgs := []DataArgs{
		{
			Value: opts.Spender,
			Type:  "address",
		},
		{
			Value: amount.String(),
			Type:  "bigInt",
		},
	}
	input := TokenTxnData{
		FunctionName: "approve",
		TokenAddress: opts.ContractAddress,
		Arguments:    dataArgs,
	}

	inputInStr, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	result := &networks.TxnBuildResult{
		Data:            string(inputInStr),
		From:            wallet.address,
		To:              opts.ContractAddress,
		Value:           big.NewInt(0),
		GasRequired:     feeInTrx,
		IsSufficientGas: wallet.balance.Cmp(feeInTrx) >= 0,
		Network:         opts.Network,

		PrivateKey: opts.PrivateKey,
	}
	return result, nil
}

func (c *TronTxnMakerClient) BuildTransferFromTxn(ctx context.Context, opts networks.TransferFromOpts) (*networks.TxnBuildResult, error) {
	wallet, err := c.getWallet(ctx, opts.PrivateKey)
	if err != nil {
		return nil, err
	}

	tokenBalance, err := c.CallTokenFunction(ctx, networks.CallTokenFunctionOpts{
		ContractAddress: opts.ContractAddress,
		FunctionName:    "balanceOf",
	}, opts.FromAddress)

	if err != nil {
		return nil, err
	}
	tokenBalanceInBigInt := new(big.Int)
	tokenBalanceInBigInt.SetString(tokenBalance, 10)
	tokenAllowance, err := c.CallTokenFunction(ctx, networks.CallTokenFunctionOpts{
		ContractAddress: opts.ContractAddress,
		FunctionName:    "allowance",
	}, opts.FromAddress, wallet.address)

	if err != nil {
		return nil, err
	}

	tokenAllowanceBigInt := new(big.Int)
	tokenAllowanceBigInt.SetString(tokenAllowance, 10)

	finalAmount := new(big.Int)
	if opts.SendAll {
		finalAmount = tokenBalanceInBigInt
	} else {
		amount, err := utils.AmountToChainUnit(opts.Amount, fmt.Sprintf("%d", opts.Decimals))
		if err != nil {
			return nil, err
		}
		finalAmount = amount
	}

	if tokenBalanceInBigInt.Cmp(finalAmount) < 0 {
		return nil, liberrors.ErrInsufficientTokenBalance
	}

	if tokenAllowanceBigInt.Cmp(finalAmount) < 0 {
		return nil, liberrors.ErrInsufficientAllowance
	}

	selector := "transferFrom(address,address,uint256)"

	params, _ := EncodeABI(
		selector,
		[]string{"address", "address", "uint256"},
		[]interface{}{
			"0x41" + TronBase58ToHex(opts.FromAddress), // hex address without base58
			"0x41" + TronBase58ToHex(opts.Destination), // hex address without base58
			finalAmount, // amount in token decimals
		},
	)

	feeInTrx, err := c.EstimateFeeDynamic(
		wallet.address,
		"",
		0,
		opts.ContractAddress,
		selector,
		params[8:], // only parameters (skip selector)
	)
	if err != nil {
		return nil, err
	}

	dataArgs := []DataArgs{
		{
			Value: opts.FromAddress,
			Type:  "address",
		},
		{
			Value: opts.Destination,
			Type:  "address",
		},
		{
			Value: finalAmount.String(),
			Type:  "bigInt",
		},
	}
	input := TokenTxnData{
		FunctionName: "transferFrom",
		TokenAddress: opts.ContractAddress,
		Arguments:    dataArgs,
	}

	inputInStr, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	result := &networks.TxnBuildResult{
		Data:            string(inputInStr),
		From:            wallet.address,
		To:              opts.ContractAddress,
		Value:           big.NewInt(0),
		GasRequired:     feeInTrx,
		IsSufficientGas: wallet.balance.Cmp(feeInTrx) >= 0,
		Network:         opts.Network,

		PrivateKey: opts.PrivateKey,
	}
	return result, nil
}

// --------------------------- Public: TRC20 - Constant ---------------------------

// CallTokenFunction calls standard TRC20 read-only methods.
// Returns decoded string/int/bignum as string where reasonable.
func (c *TronTxnMakerClient) CallTokenFunction(ctx context.Context, opts networks.CallTokenFunctionOpts, args ...interface{}) (string, error) {

	contract41, err := tronAddressToHex(opts.ContractAddress)
	if err != nil {
		return "", fmt.Errorf("invalid contract address: %v", err)
	}

	owner41, err := tronAddressToHex("TWaWiCyKbtZR1rCsPmzmjBtjNJGVKgvFjp")
	if err != nil {
		return "", fmt.Errorf("invalid owner address: %v", err)
	}
	var selector, param string
	switch opts.FunctionName {
	case "name":
		selector = "name()"
		param = ""
	case "symbol":
		selector = "symbol()"
		param = ""
	case "decimals":
		selector = "decimals()"
		param = ""
	case "balanceOf":
		if len(args) < 1 {
			return "", errors.New("balanceOf requires 1 argument")
		}
		addr41, err := tronAddressToHex(asString(args[0]))
		if err != nil {
			return "", fmt.Errorf("invalid address arg: %v", err)
		}
		selector = "balanceOf(address)"
		param = leftPad64(addr41[2:]) // 20 bytes (no 41), left-padded
	case "allowance":
		if len(args) < 2 {
			return "", errors.New("allowance requires 2 arguments")
		}
		ownerArg41, err := tronAddressToHex(asString(args[0]))
		if err != nil {
			return "", fmt.Errorf("invalid owner arg: %v", err)
		}
		spenderArg41, err := tronAddressToHex(asString(args[1]))
		if err != nil {
			return "", fmt.Errorf("invalid spender arg: %v", err)
		}
		selector = "allowance(address,address)"
		param = leftPad64(ownerArg41[2:]) + leftPad64(spenderArg41[2:])
	default:
		return "", fmt.Errorf("unsupported function: %s", opts.FunctionName)
	}

	payload := map[string]interface{}{
		"contract_address":  contract41,
		"function_selector": selector,
		"parameter":         param, // only params (no selector)
		"visible":           false,
	}
	if owner41 != "" {
		payload["owner_address"] = owner41
	}

	res, err := c.httpPost("/wallet/triggerconstantcontract", payload)
	if err != nil {
		return "", fmt.Errorf("failed to call contract: %v", err)
	}

	// Parse result
	constRes, ok := res["constant_result"].([]interface{})
	if !ok || len(constRes) == 0 {
		return "", fmt.Errorf("empty result from contract call: %v", res["result"])
	}
	hexOut, _ := constRes[0].(string)
	if hexOut == "" {
		return "", errors.New("contract returned empty hex result")
	}

	switch opts.FunctionName {
	case "name", "symbol":
		str, derr := decodeTronString(hexOut)
		if derr != nil {
			// Fallback: return raw hex
			return hexOut, nil
		}
		return str, nil
	case "decimals":
		n := new(big.Int)
		n.SetString(hexOut, 16)
		return n.String(), nil
	case "balanceOf", "allowance":
		n := new(big.Int)
		n.SetString(hexOut, 16)
		return n.String(), nil
	default:
		return hexOut, nil
	}
}

func (c *TronTxnMakerClient) EstimateFeeDynamic(owner, to string, amountSun int64, contractAddress, functionSelector, parameter string) (*big.Int, error) {

	// If no contract address => normal TRX transfer
	if contractAddress == "" {
		payload := map[string]interface{}{
			"owner_address": owner,
			"to_address":    to,
			"amount":        amountSun,
			"visible":       true,
		}

		resp, err := c.httpPost("/wallet/createtransaction", payload)
		if err != nil {
			return nil, fmt.Errorf("failed TRX transfer estimate: %v", err)
		}
		jss, err := json.Marshal(resp)
		if err != nil {
			return nil, fmt.Errorf("failed TRX transfer estimate: %v", err)
		}
		fmt.Println("netUsageRawnetUsageRaw", string(jss))

		netUsageRaw := resp["net_usage"]
		if netUsageRaw == nil {
			return nil, nil // free transfer (enough bandwidth)
		}

		netUsage := int64(netUsageRaw.(float64))

		return big.NewInt(netUsage), nil
	}

	// Otherwise: smart contract call (TRC20 / custom contract)
	payload := map[string]interface{}{
		"owner_address":     owner,
		"contract_address":  contractAddress,
		"function_selector": functionSelector,
		"parameter":         parameter,
		"visible":           true,
	}

	resp, err := c.httpPost("/wallet/triggerconstantcontract", payload)
	if err != nil {
		return nil, fmt.Errorf("failed contract estimate: %v", err)
	}

	energyUsed := int64(resp["energy_used"].(float64))

	feeSun := energyUsed * 280 // SUN cost

	return big.NewInt(feeSun), nil
}

func (c *TronTxnMakerClient) BroadcastTxn(ctx context.Context, txn *networks.TxnBuildResult) (string, error) {

	if !txn.IsSufficientGas {
		return "", liberrors.ErrInsufficientBalance
	}
	wallet, err := c.getWallet(ctx, txn.PrivateKey)
	if err != nil {
		return "", err
	}

	if wallet.balance.Cmp(big.NewInt(0).Add(txn.Value, txn.GasRequired)) < 0 {
		return "", liberrors.ErrInsufficientBalance
	}
	dest, err := address.FromBase58(txn.To)
	if err != nil {
		return "", err
	}
	if txn.Data == "" {
		txInfo, err := c.trxClient.Transfer(wallet.account, dest, txn.Value.Uint64())
		if err != nil {
			return "", err
		}
		return txInfo.Id, nil
	} else {

		type TxnResult struct {
			Success bool `abi:"success"`
		}

		var contractInput TokenTxnData
		err := json.Unmarshal([]byte(txn.Data), &contractInput)
		if err != nil {
			return "", err
		}

		sender, err := tronWallet.FromPrivateKeyHex(txn.PrivateKey)
		if err != nil {
			return "", err
		}

		tokenContract, err := address.FromBase58(contractInput.TokenAddress)
		if err != nil {
			return "", err
		}
		var txnResult TxnResult

		args := make([]interface{}, len(contractInput.Arguments))
		for i, arg := range contractInput.Arguments {
			switch arg.Type {
			case "address":
				value, err := address.FromBase58(arg.Value)
				if err != nil {
					return "", err
				}
				args[i] = value
			case "bigInt":
				value, _ := big.NewInt(0).SetString(arg.Value, 10)
				args[i] = value
			default:
				return "", fmt.Errorf("unsupported argument type: %s", arg.Type)
			}
		}
		txnInfo, err := c.trxClient.CallContract(sender, tronClient.CallContractInput{
			Address:   tokenContract,
			Function:  erc20ABI.Functions[contractInput.FunctionName],
			Arguments: args,
			FeeLimit:  10_000_000,
			CallValue: 0,
			Result:    &txnResult,
		})
		if err != nil {
			return "", err
		}

		return txnInfo.Id, nil

	}

}

func (c *TronTxnMakerClient) GetNativeBalance(ctx context.Context, opts networks.NativeBalanceOpts) (*big.Int, error) {
	payload := map[string]interface{}{
		"address": opts.Address,
		"visible": true, // so we can use base58 address
	}

	// Call Tron node API
	resp, err := c.httpPost("/walletsolidity/getaccount", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account: %v", err)
	}

	// If account doesn’t exist => balance = 0
	balanceRaw, ok := resp["balance"]
	if !ok {
		return big.NewInt(0), nil
	}

	// Tron returns balance in SUN (int)
	balanceSun, ok := balanceRaw.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid balance field type: %T", balanceRaw)
	}

	// Convert SUN -> TRX
	balanceTrx := balanceSun
	return big.NewInt(int64(balanceTrx)), nil
}

func (c *TronTxnMakerClient) getWallet(ctx context.Context, privateKey string) (wallet *TronWallet, err error) {

	account, err := tronWallet.FromPrivateKeyHex(privateKey)
	if err != nil {
		return nil, err
	}
	balance, err := c.GetNativeBalance(ctx, networks.NativeBalanceOpts{
		Address: account.Address().ToBase58(),
	})
	if err != nil {
		return nil, err
	}
	return &TronWallet{
		account: account,
		address: account.Address().ToBase58(),
		balance: balance,
	}, nil
}

// --------------------------- HTTP + Signing ---------------------------

func (c *TronTxnMakerClient) httpPost(endpoint string, data interface{}) (map[string]interface{}, error) {
	url := c.Rpc + endpoint

	var bodyReader io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %v", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	} else {
		bodyReader = bytes.NewBufferString(`{}`)
	}

	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Helpful for debugging:
	// fmt.Printf("%s -> %s\n", endpoint, string(raw))
	return result, nil
}

// --------------------------- Private: TRON ---------------------------
// tronAddressToHex converts Base58/0x.. or 41.. to 41.. (no 0x).
func tronAddressToHex(addr string) (string, error) {
	if addr == "" {
		return "", errors.New("empty address")
	}
	a := strings.TrimSpace(addr)

	// Already 41.. hex (no 0x)
	if len(a) == 42 && strings.HasPrefix(a, "41") {
		return a, nil
	}

	// 0x.. (Ethereum-like 20 bytes)
	if strings.HasPrefix(a, "0x") && len(a) == 42 {
		return "41" + a[2:], nil
	}

	// Base58 (T...)
	decoded := base58.Decode(a)
	if len(decoded) < 25 {
		return "", errors.New("invalid base58 address")
	}
	// Verify checksum
	payload := decoded[:len(decoded)-4]
	cs := decoded[len(decoded)-4:]
	h1 := sha256.Sum256(payload)
	h2 := sha256.Sum256(h1[:])
	for i := 0; i < 4; i++ {
		if cs[i] != h2[i] {
			return "", errors.New("invalid base58 checksum")
		}
	}
	// payload[0] should be 0x41. Final hex is 41 + 20 bytes.
	if len(payload) != 21 || payload[0] != 0x41 {
		return "", errors.New("invalid base58 payload for tron")
	}
	return hex.EncodeToString(payload), nil
}

// PrivateKeyHexToTronAddress returns (Base58, 0x41.. hex).
func PrivateKeyHexToTronAddress(privHex string) (string, string, error) {
	keyBytes, err := hex.DecodeString(strings.TrimPrefix(privHex, "0x"))
	if err != nil {
		return "", "", fmt.Errorf("decode priv: %w", err)
	}
	if len(keyBytes) != 32 {
		return "", "", errors.New("private key must be 32 bytes")
	}
	sk, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return "", "", fmt.Errorf("to ecdsa: %w", err)
	}
	ethAddr := crypto.PubkeyToAddress(sk.PublicKey)        // 20 bytes
	trxPayload := append([]byte{0x41}, ethAddr.Bytes()...) // 21 bytes
	// base58 with double-SHA256 checksum (first 4 bytes)
	h1 := sha256.Sum256(trxPayload)
	h2 := sha256.Sum256(h1[:])
	b58 := base58.Encode(append(trxPayload, h2[:4]...))
	return b58, "0x" + hex.EncodeToString(trxPayload), nil
}

// decodeTronString decodes ABI-encoded string output.
func decodeTronString(hexResult string) (string, error) {
	data, err := hex.DecodeString(hexResult)
	if err != nil {
		return "", fmt.Errorf("invalid hex: %w", err)
	}
	if len(data) < 64 {
		return "", errors.New("too short to decode ABI string")
	}
	// offset := data[0:32] (ignored)
	length := new(big.Int).SetBytes(data[32:64]).Int64()
	if length < 0 || int(64+length) > len(data) {
		return "", fmt.Errorf("invalid string length %d > data size %d", length, len(data))
	}
	return string(data[64 : 64+length]), nil
}

// leftPad64 pads a hex string (no 0x) to 32 bytes (64 hex chars).
func leftPad64(hexNoPrefix string) string {
	return strings.Repeat("0", 64-len(hexNoPrefix)) + hexNoPrefix
}

func asString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
