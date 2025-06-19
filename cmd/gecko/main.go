package main

import (
	"log"
	"os"

	"github.com/jkeddari/walletscan/internal/gecko"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	coinGeckoAPIKey := os.Getenv("COINGECKO_APIKEY")
	if coinGeckoAPIKey == "" {
		log.Fatal("missing coingecko api key")
	}

	err = gecko.FetchCoins(coinGeckoAPIKey, "./coinlist.json", 100)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("file write on ./coinlist.json")
}
