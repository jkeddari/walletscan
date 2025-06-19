package walletscan

import (
	"errors"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jkeddari/walletscan/internal/bitcoin"
	"github.com/jkeddari/walletscan/internal/evmscan"
	"github.com/jkeddari/walletscan/internal/gecko"
	"github.com/jkeddari/walletscan/internal/solana"
	"github.com/jkeddari/walletscan/internal/tronscan"
	"github.com/jkeddari/walletscan/internal/types"
)

//go:generate go tool templ generate

var (
	Logger           = slog.Default()
	COINGECKO_APIKEY string
)

func Scan(address string, coinsDetails *gecko.CoinsDetails, prices gecko.Prices) (*types.WalletData, error) {
	switch IsAddress(address) {
	case types.EVMAddress:
		return evmscan.Scan(address, coinsDetails, prices)
	case types.BitcoinAddress:
		return bitcoin.Scan(address, coinsDetails.Icon("bitcoin"), prices.PriceByID("bitcoin"))
	case types.SolanaAddress:
		return solana.Scan(address, coinsDetails, prices)
	case types.TronAddress:
		return tronscan.Scan(address)
	}
	return nil, errors.New("bad address")
}

func IsAddress(address string) types.AddressType {
	switch true {
	case common.IsHexAddress(address):
		return types.EVMAddress
	case bitcoin.ValidAddress(address):
		return types.BitcoinAddress
	case solana.ValidAddress(address):
		return types.SolanaAddress
	case tronscan.ValidAddress(address):
		return types.TronAddress
	default:
		return types.UnknownAddress
	}
}
