package types

type Network string

const (
	BTC     Network = "BTC"
	ETH     Network = "ETH"
	POLYGON Network = "POLYGON"
	BSC     Network = "BSC"
	TRON    Network = "TRON"
	SOL     Network = "SOL"

	BTC_TESTNET  Network = "BTC_TESTNET"
	SEPOLIA      Network = "SEPOLIA"
	POLYGON_AMOY Network = "AMOY"
	BSC_TESTNET  Network = "BSC_TESTNET"
	SHASTA       Network = "SHASTA"
	SOL_TESTNET  Network = "SOL_TESTNET"
)

var EVMNetworks = []Network{
	ETH,
	POLYGON,
	BSC,
	SEPOLIA,
	POLYGON_AMOY,
	BSC_TESTNET,
}
