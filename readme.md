# Go Crypto Sender

[![Go Reference](https://pkg.go.dev/badge/github.com/prozeb/go-crypto-sender.svg)](https://pkg.go.dev/github.com/prozeb/go-crypto-sender)
[![Go Report Card](https://goreportcard.com/badge/github.com/prozeb/go-crypto-sender)](https://goreportcard.com/report/github.com/prozeb/go-crypto-sender)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go Crypto Sender is a versatile Go library that simplifies cryptocurrency transactions across multiple blockchain networks. It provides a unified interface for sending native tokens, transferring tokens, and interacting with smart contracts on various blockchains.

## Features

- **Multi-Network Support**: Works with multiple blockchain networks including:
  - EVM-compatible chains (Ethereum, Polygon, BSC, etc.)
  - TRON (Mainnet & Shasta testnet)
  - Bitcoin (Mainnet & Testnet)
- **Transaction Types**:
  - Native token transfers
  - ERC-20/TRC-20 token transfers
  - Token approvals
  - Transfer from (for pre-approved tokens)
  - Custom smart contract function calls
- **Simple API**: Clean and intuitive interface for all operations
- **Thread-Safe**: Designed for concurrent use

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
)

func main() {
    // Initialize client with RPC endpoints
    rpcs := map[types.Network]string{
        types.ETHEREUM: "https://mainnet.infura.io/v3/YOUR-API-KEY",
        types.TRON:     "https://api.trongrid.io",
        types.BTC:      "https://btc-node.example.com",
    }

    client, err := gocryptosender.NewTxnMakerClient(rpcs)
    if err != nil {
        panic(err)
    }

    ctx := context.Background()

    // Example: Send native token
    txHash, err := client.MakeNativeTxn(ctx, networks.NativeTxnOpts{
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
    fmt.Printf("Transaction sent: %s\n", txHash)
}
```

## Supported Networks

The library supports the following networks (defined in `types.Network`):

- `ETHEREUM`
- `POLYGON`
- `BSC`
- `TRON`
- `SHASTA` (TRON testnet)
- `BTC`
- `BTC_TESTNET`

## API Reference

### Initialize Client

```go
func NewTxnMakerClient(rpcs map[types.Network]string) (*TxnMakerClient, error)
```

### Available Methods

1. **Native Token Transfer**
   ```go
   MakeNativeTxn(ctx context.Context, opts networks.NativeTxnOpts) (string, error)
   ```

2. **Token Transfer**
   ```go
   TransferToken(ctx context.Context, opts networks.TransferTokenOpts) (string, error)
   ```

3. **Token Approval**
   ```go
   ApproveToken(ctx context.Context, opts networks.ApproveTokenOpts) (string, error)
   ```

4. **Transfer From**
   ```go
   TransferFrom(ctx context.Context, opts networks.TransferFromOpts) (string, error)
   ```

5. **Call Token Function**
   ```go
   CallTokenFunction(ctx context.Context, opts networks.CallTokenFunctionOpts, args ...interface{}) (string, error)
   ```

## Error Handling

The library defines custom error types in the `liberrors` package. Always check and handle errors returned by the library functions.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

This software is provided "as is" and any expressed or implied warranties, including, but not limited to, the implied warranties of merchantability and fitness for a particular purpose are disclaimed. In no event shall the authors or copyright holders be liable for any direct, indirect, incidental, special, exemplary, or consequential damages.
```

### Step 3: After adding the README, commit and push your changes

```bash
git add readme.md
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

Would you like me to help you with anything else?