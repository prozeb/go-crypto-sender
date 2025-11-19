# Go Crypto Sender

[![Go Reference](https://pkg.go.dev/badge/github.com/prozeb/go-crypto-sender.svg)](https://pkg.go.dev/github.com/prozeb/go-crypto-sender)
[![Go Report Card](https://goreportcard.com/badge/github.com/prozeb/go-crypto-sender)](https://goreportcard.com/report/github.com/prozeb/go-crypto-sender)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go Crypto Sender is a powerful Go library that provides a unified interface for creating, signing, and broadcasting transactions across multiple blockchain networks.

## Features

- **Multi-Network Support**:
  - EVM-compatible chains (Ethereum, Polygon, BSC)
  - TRON (Mainnet & Shasta testnet)
  - Bitcoin (Mainnet & Testnet)

- **Core Functionality**:
  - Native token transfers
  - ERC-20/TRC-20 token transfers
  - Token approvals
  - Transfer from (for pre-approved tokens)
  - Custom smart contract interactions
  - Balance queries

## Installation

```bash
go get github.com/prozeb/go-crypto-sender
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/prozeb/go-crypto-sender"
    "github.com/prozeb/go-crypto-sender/types"
    "github.com/prozeb/go-crypto-sender/networks"
)

func main() {
    // Initialize client with RPC endpoints
    rpcs := map[types.Network]string{
        types.ETHEREUM: "https://mainnet.infura.io/v3/YOUR-API-KEY",
        types.TRON:     "https://api.trongrid.io",
    }

    client, err := gocryptosender.NewTxnMakerClient(rpcs)
    if err != nil {
        panic(err)
    }

    ctx := context.Background()

    // Example: Transfer native tokens
    result, err := client.BuildTransferNativeTxn(ctx, networks.NativeTxnOpts{
        Network:   types.ETHEREUM,
        From:      "0x...",
        To:        "0x...",
        Amount:    "1000000000000000000", // 1 ETH in wei
        GasLimit:  21000,
        GasPrice:  "20000000000", // 20 Gwei
        PrivateKey: "0x...",
    })
    if err != nil {
        panic(err)
    }

    // Broadcast the transaction
    txHash, err := client.BroadcastTxn(ctx, result)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Transaction sent: %s\n", txHash)
}
```

## API Reference

### Initialize Client
```go
func NewTxnMakerClient(rpcs map[types.Network]string) (*TxnMakerClient, error)
```

### Available Methods

1. **Build and Sign Native Token Transfer**
   ```go
   BuildTransferNativeTxn(ctx context.Context, opts networks.NativeTxnOpts) (*networks.TxnBuildResult, error)
   ```

2. **Build and Sign Token Transfer**
   ```go
   BuildTransferTokenTxn(ctx context.Context, opts networks.TransferTokenOpts) (*networks.TxnBuildResult, error)
   ```

3. **Get Native Balance**
   ```go
   GetNativeBalance(ctx context.Context, opts networks.NativeBalanceOpts) (*big.Int, error)
   ```

4. **Build Token Approval**
   ```go
   BuildApproveTokenTxn(ctx context.Context, opts networks.ApproveTokenOpts) (*networks.TxnBuildResult, error)
   ```

5. **Broadcast Transaction**
   ```go
   BroadcastTxn(ctx context.Context, txn *networks.TxnBuildResult) (string, error)
   ```

6. **Build Transfer From**
   ```go
   BuildTransferFromTxn(ctx context.Context, opts networks.TransferFromOpts) (*networks.TxnBuildResult, error)
   ```

7. **Call Token Function**
   ```go
   CallTokenFunction(ctx context.Context, opts networks.CallTokenFunctionOpts, args ...interface{}) (string, error)
   ```

## Error Handling

The library uses standard Go error handling. All methods return an error that should be checked.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT

## Disclaimer

This software is provided "as is" without any warranties. Use at your own risk.
git commit -m "Add comprehensive README.md"
git push origin main
```

The README includes:
- Project badges for GoDoc and Go Report Card
- A clear description of the library's purpose
- Key features
- Installation instructions
- A complete quick start example
- List of supported networks
- Detailed API reference
- Error handling information
- Contributing guidelines
- License information
- Standard disclaimer

