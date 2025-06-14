package evmscan

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

const erc20ABI = `[
  {
    "constant": true,
    "inputs": [{ "name": "account", "type": "address" }],
    "name": "balanceOf",
    "outputs": [{ "name": "", "type": "uint256" }],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "constant": true,
    "inputs": [],
    "name": "decimals",
    "outputs": [{ "name": "", "type": "uint8" }],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "constant": true,
    "inputs": [],
    "name": "name",
    "outputs": [{ "name": "", "type": "string" }],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "constant": true,
    "inputs": [],
    "name": "symbol",
    "outputs": [{ "name": "", "type": "string" }],
    "stateMutability": "view",
    "type": "function"
  }
]`

var networks = []*networkInfo{
	{
		Network:  "ethereum",
		URL:      "https://ethereum-rpc.publicnode.com",
		SymbolID: "ethereum",
	},
	{
		Network:  "binance-smart-chain",
		URL:      "https://bsc-rpc.publicnode.com",
		SymbolID: "binancecoin",
	},
	{
		Network:  "polygon",
		URL:      "https://polygon-bor-rpc.publicnode.com",
		SymbolID: "polygon-ecosystem-token",
	},
	{
		Network:  "base",
		URL:      "https://base-rpc.publicnode.com",
		SymbolID: "ethereum",
	},
	{
		Network:  "avalanche",
		URL:      "https://avalanche-c-chain-rpc.publicnode.com",
		SymbolID: "avalanche-2",
	},
	{
		Network:  "arbitrum",
		URL:      "https://arbitrum-one-rpc.publicnode.com",
		SymbolID: "ethereum",
	},
	{
		Network:  "optimism",
		URL:      "https://optimism-rpc.publicnode.com",
		SymbolID: "ethereum",
	},
}

type networkInfo struct {
	Network  string
	URL      string
	SymbolID string
}

func weiToEther(wei *big.Int) float64 {
	bfloat, _ := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether)).Float64()
	return bfloat
}

func parseTokenAmount(balance *big.Int, decimals int) float64 {
	balanceFloat := new(big.Float).SetInt(balance)
	divisor := new(big.Float).SetFloat64(1)
	for i := 0; i < decimals; i++ {
		divisor.Mul(divisor, big.NewFloat(10))
	}
	humanValue := new(big.Float).Quo(balanceFloat, divisor)
	v, _ := humanValue.Float64()
	return v
}
