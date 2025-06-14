package gecko

type PlatformDetails struct {
	Decimal  int    `json:"decimal_place"`
	Contract string `json:"contract_address"`
}

// Coin defines Coin details from CoinGecko.
type Coin struct {
	Name       string                     `json:"name"`
	Symbol     string                     `json:"symbol"`
	Plateforms map[string]PlatformDetails `json:"detail_platforms"`
}

type CoinsDetails struct {
	IDs   []string
	Coins map[string]*Coin
}

func (c *CoinsDetails) Name(id string) string {
	if coin, ok := c.Coins[id]; ok {
		return coin.Name
	}

	return "unknown"
}

func (c *CoinsDetails) Symbol(id string) string {
	if coin, ok := c.Coins[id]; ok {
		return coin.Symbol
	}

	return "unknown"
}

func (c *CoinsDetails) CoinByContract(network, contract string) (string, *Coin) {
	for id, detail := range c.Coins {
		if tokenContract, ok := detail.Plateforms[network]; ok {
			if tokenContract.Contract == contract {
				return id, detail
			}
		}
	}
	return "", nil
}

// Price defines prices by token ID.
type Prices map[string]float64

func (p Prices) PriceByID(id string) float64 {
	if price, ok := p[id]; ok {
		return price
	}

	return -1
}
