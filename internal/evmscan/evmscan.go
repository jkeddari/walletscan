package evmscan

import (
	"context"
	"errors"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jkeddari/walletscan/internal/gecko"
	"github.com/jkeddari/walletscan/internal/types"
)

const maxRetries = 3

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

	var wg sync.WaitGroup
	for _, network := range networks {
		wg.Add(1)
		go func(n *networkInfo) {
			defer wg.Done()

			logger.Debug("scanning network", "network", n)
			if err := ctx.networkScan(n); err != nil {
				logger.Error("scan network failed", "network", n.Network, "err", err)
			}
			logger.Debug("scanning network done", "network", n)
		}(network)
	}

	wg.Wait()
	return ctx.walletData, nil
}

func (c *scanContext) networkScan(network *networkInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, network.URL)
	if err != nil {
		return err
	}

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
		time.Sleep(5 * time.Millisecond)

		if coinDetails, ok := coin.Plateforms[network.Network]; ok {
			var amount float64
			var err error

			for attempt := 1; attempt <= maxRetries; attempt++ {
				amount, err = c.balanceOf(client, coinDetails)
				if err == nil {
					break
				}

				logger.Warn("balanceOf retry failed", "attempt", attempt, "network", network.Network, "id", id, "error", err)
				time.Sleep(time.Duration(100*attempt) * time.Millisecond)
			}

			if err != nil {
				logger.Error("token scan failed", "network", network.Network, "id", id, "error", err)
				continue
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := client.CallContract(ctx, msg, nil)
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
