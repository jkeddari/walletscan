package bitcoin

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/jkeddari/walletscan/internal/types"
)

const urlBlockInfo = "https://blockchain.info/rawaddr/"

func ValidAddress(address string) bool {
	addr, err := btcutil.DecodeAddress(address, &chaincfg.MainNetParams)
	if err != nil {
		return false
	}
	return addr.IsForNet(&chaincfg.MainNetParams)
}

func Scan(address, iconURL string, bitcoinPrice float64) (*types.WalletData, error) {
	if !ValidAddress(address) {
		return nil, errors.New("bad address")
	}

	url := urlBlockInfo + address
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info struct {
		Balance int `json:"final_balance"`
	}

	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return nil, err
	}

	amount := float64(info.Balance) / 1e8
	return &types.WalletData{
		TotalValue: amount * bitcoinPrice,
		Network: map[string]struct{}{
			"bitcoin": {},
		},
		Balances: map[string]*types.Balance{
			"bitcoin": {
				TokenID: "bitcoin",
				Amount: map[string]float64{
					"bitcoin": amount,
				},
				Symbol:      "BTC",
				Name:        "Bitcoin",
				Price:       bitcoinPrice,
				TotalAmount: amount,
				IconURL:     iconURL,
			},
		},
	}, nil
}
