package evmscan

import (
	"context"
	"errors"
	"log/slog"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jkeddari/walletscan/internal/gecko"
	"github.com/jkeddari/walletscan/internal/types"
)

var logger = slog.Default()

type scanContext struct {
	address      common.Address
	coinsDetails *gecko.CoinsDetails
	prices       gecko.Prices
	erc20ABI     abi.ABI
	walletData   *types.WalletData
}

func Scan(address string, coinsDetails *gecko.CoinsDetails, prices gecko.Prices) (*types.WalletData, error) {
	if !common.IsHexAddress(address) {
		return nil, errors.New("bad address")
	}

	if coinsDetails == nil || prices == nil {
		return nil, errors.New("bad gecko details")
	}

	evmAddress := common.HexToAddress(address)
	logger.Info("scan evm address", "address", evmAddress.String())

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, err
	}

	ctx := scanContext{
		address:      evmAddress,
		coinsDetails: coinsDetails,
		prices:       prices,
		erc20ABI:     parsedABI,
		walletData:   types.NewWalletData(),
	}

	for _, network := range networks {
		logger.Debug("scanning network", "network", network)
		if err := ctx.networkScan(network); err != nil {
			logger.Error("scan network failed", "network", network.Network)
		}

	}

	return ctx.walletData, nil
}

func (c *scanContext) networkScan(network *networkInfo) error {
	client, err := ethclient.Dial(network.URL)
	if err != nil {
		return err
	}

	ctx := context.Background()

	block, err := client.BlockNumber(ctx)
	if err != nil {
		return err
	}

	weiAmount, err := client.BalanceAt(ctx, c.address, big.NewInt(int64(block)))
	if err != nil {
		return err
	}

	c.walletData.Add(
		network.SymbolID,
		network.Network,
		c.coinsDetails.Symbol(network.SymbolID),
		c.coinsDetails.Name(network.SymbolID),
		weiToEther(weiAmount),
		c.prices.PriceByID(network.SymbolID),
	)

	for id, coin := range c.coinsDetails.Coins {
		if coinDetails, ok := coin.Plateforms[network.Network]; ok {
			amount, err := c.balanceOf(client, coinDetails)
			if err != nil {
				logger.Error("token scan failed", "network", network.Network, "id", id)
			}
			c.walletData.Add(
				id,
				network.Network,
				c.coinsDetails.Symbol(id),
				c.coinsDetails.Name(id),
				amount,
				c.prices.PriceByID(id),
			)
		}
	}

	return nil
}

func (c *scanContext) balanceOf(client *ethclient.Client, tokenDetail gecko.PlatformDetails) (float64, error) {
	contract := common.HexToAddress(tokenDetail.Contract)

	data, err := c.erc20ABI.Pack("balanceOf", c.address)
	if err != nil {
		return 0, err
	}
	msg := ethereum.CallMsg{To: &contract, Data: data}
	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return 0, err
	}
	var out *big.Int
	err = c.erc20ABI.UnpackIntoInterface(&out, "balanceOf", result)
	if err != nil {
		return 0, err
	}

	return parseTokenAmount(out, tokenDetail.Decimal), nil
}
