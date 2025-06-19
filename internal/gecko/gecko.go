package gecko

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultCoinGeckoURL = "https://api.coingecko.com/api/v3/"

func ReadCoinsDetails(file string) (*CoinsDetails, error) {
	details := CoinsDetails{
		Coins: make(map[string]*Coin),
	}

	payload, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(payload, &details.Coins); err != nil {
		return nil, err
	}

	for id := range details.Coins {
		details.IDs = append(details.IDs, id)
	}

	return &details, nil
}

// Prices returns current price per coin ID.
func PricesByIDs(apiKey string, ids ...string) (Prices, error) {
	baseURL := defaultCoinGeckoURL + "coins/markets"
	params := url.Values{}
	params.Set("vs_currency", "usd")
	params.Set("precision", "full")
	params.Set("ids", strings.Join(ids, ","))

	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("x-cg-demo-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []struct {
		ID           string  `json:"id"`
		CurrentPrice float64 `json:"current_price"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	prices := make(Prices)
	for _, coin := range result {
		prices[coin.ID] = coin.CurrentPrice
	}

	return prices, nil
}

// FetchCoins fetch n first top coins from CoinGecko.
func FetchCoins(apiKey, destFile string, n int) error {
	topCoins, err := getTopCoinIDs(apiKey, n)
	if err != nil {
		return err
	}

	coinsList := map[string]*Coin{}
	for _, id := range topCoins {
		coin, err := getCoinWithPlatforms(apiKey, id)
		if err != nil {
			// logger.Error("fetch coin", "id", id, "err", err)
			continue
		}
		coinsList[id] = coin
		time.Sleep(2000 * time.Millisecond) // limit rate (200/min)
	}

	payload, err := json.MarshalIndent(coinsList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(destFile, payload, 0o644)
}

func getTopCoinIDs(apiKey string, n int) ([]string, error) {
	var ids []string
	perPage := 250
	page := 1

	for len(ids) < n {
		url := fmt.Sprintf(
			defaultCoinGeckoURL+"coins/markets?vs_currency=usd&order=market_cap_desc&per_page=%d&page=%d",
			perPage, page,
		)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Add("accept", "application/json")
		req.Header.Add("x-cg-demo-api-key", apiKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("%s", resp.Status)
		}

		var res []struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return nil, err
		}

		if len(res) == 0 {
			break
		}

		for _, coin := range res {
			ids = append(ids, coin.ID)
			if len(ids) >= n {
				break
			}
		}

		page++
	}

	return ids, nil
}

func getCoinWithPlatforms(apiKey, id string) (*Coin, error) {
	url := defaultCoinGeckoURL + "coins/" + id

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("x-cg-demo-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		ID         string                     `json:"id"`
		Name       string                     `json:"name"`
		Symbol     string                     `json:"symbol"`
		Plateforms map[string]PlatformDetails `json:"detail_platforms"`
		Image      struct {
			Small string `json:"small"`
		} `json:"image"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &Coin{
		Name:       data.Name,
		Symbol:     data.Symbol,
		Plateforms: data.Plateforms,
		Icon:       data.Image.Small,
	}, nil
}
