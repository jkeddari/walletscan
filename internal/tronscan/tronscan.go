package tronscan

import (
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/url"
	"strconv"

	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/jkeddari/walletscan/internal/types"
)

const tronscanURL = "https://apilist.tronscanapi.com/api/account/token_asset_overview"

type tokenData struct {
	TokenID         string  `json:"tokenId"`
	TokenName       string  `json:"tokenName"`
	TokenAbbr       string  `json:"tokenAbbr"`
	TokenDecimal    int     `json:"tokenDecimal"`
	TokenCanShow    int     `json:"tokenCanShow"`
	TokenType       string  `json:"tokenType"`
	TokenLogo       string  `json:"tokenLogo"`
	VIP             bool    `json:"vip"`
	Balance         string  `json:"balance"`
	TokenPriceInTrx float64 `json:"tokenPriceInTrx"`
	TokenPriceInUsd float64 `json:"tokenPriceInUsd"`
	AssetInTrx      float64 `json:"assetInTrx"`
	AssetInUsd      float64 `json:"assetInUsd"`
	Percent         float64 `json:"percent"`
}

type refreshTimeInfo struct {
	Type           string `json:"type"`
	LastUpdateTime int64  `json:"lastUpdateTime"`
}

type tronWalletResponse struct {
	TotalAssetInTrx float64         `json:"totalAssetInTrx"`
	Data            []tokenData     `json:"data"`
	TotalTokenCount int             `json:"totalTokenCount"`
	RefreshTimeInfo refreshTimeInfo `json:"refreshTimeInfo"`
	TotalAssetInUsd float64         `json:"totalAssetInUsd"`
}

func convertBalance(balanceStr string, decimals int) (float64, error) {
	bigIntBalance, ok := new(big.Int).SetString(balanceStr, 10)
	if !ok {
		return 0, strconv.ErrSyntax
	}

	pow := new(big.Float).SetFloat64(1)
	for i := 0; i < decimals; i++ {
		pow.Mul(pow, big.NewFloat(10))
	}

	balanceFloat := new(big.Float).SetInt(bigIntBalance)

	adjusted := new(big.Float).Quo(balanceFloat, pow)

	result, _ := adjusted.Float64()
	return result, nil
}

func Scan(address string) (*types.WalletData, error) {
	if !ValidAddress(address) {
		return nil, errors.New("bad address")
	}

	params := url.Values{}
	params.Set("address", address)

	req, err := http.NewRequest("GET", tronscanURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tronWallet tronWalletResponse
	err = json.NewDecoder(resp.Body).Decode(&tronWallet)
	if err != nil {
		return nil, err
	}

	result := types.NewWalletData()
	for _, token := range tronWallet.Data {
		balance, err := convertBalance(token.Balance, token.TokenDecimal)
		if err != nil {
			return nil, err
		}
		result.Add(token.TokenID, "tron", token.TokenAbbr, token.TokenName, token.TokenLogo, balance, token.TokenPriceInUsd)
	}

	return result, nil
}

func ValidAddress(addr string) bool {
	if _, err := address.Base58ToAddress(addr); err == nil {
		return true
	}
	return false
}
