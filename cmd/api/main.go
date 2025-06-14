package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/jkeddari/walletscan"
	"github.com/jkeddari/walletscan/internal/gecko"
	"github.com/jkeddari/walletscan/ui/pages"
	"github.com/joho/godotenv"
)

const defaultPort = "8090"

var (
	coinsInfo       *gecko.CoinsDetails
	coinGeckoAPIKey string
	logger          = slog.Default()
)

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	coinGeckoAPIKey = os.Getenv("COINGECKO_APIKEY")
	if coinGeckoAPIKey == "" {
		log.Fatal("missing coingecko api key")
	}

	coinsInfo, err = gecko.ReadCoinsDetails("./coinlist.json")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	mux := http.NewServeMux()
	setupAssetsRoutes(mux)
	mux.Handle("GET /", templ.Handler(pages.Landing()))
	mux.Handle("GET /about", templ.Handler(pages.About()))
	mux.HandleFunc("POST /scan-address", func(w http.ResponseWriter, r *http.Request) {
		address := r.FormValue("address")
		if address == "" {
			http.Error(w, "Address required", http.StatusBadRequest)
			return
		}

		// HTMX redirection
		w.Header().Set("HX-Redirect", "/scan/"+address)
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /scan/{address}", func(w http.ResponseWriter, r *http.Request) {
		address := r.PathValue("address")
		if address == "" {
			http.Error(w, "Address required", http.StatusBadRequest)
			return
		}

		prices, err := gecko.PricesByIDs(coinGeckoAPIKey, coinsInfo.IDs...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data, err := walletscan.Scan(address, coinsInfo, prices)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		pages.Scan(address, data).Render(context.Background(), w)
	})

	logger.Info("Server is running", "port", port, "coins", len(coinsInfo.IDs))
	http.ListenAndServe(":"+port, mux)
}

func setupAssetsRoutes(mux *http.ServeMux) {
	// isDevelopment := os.Getenv("GO_ENV") != "production"

	assetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if isDevelopment {
		// 	w.Header().Set("Cache-Control", "no-store")
		// }

		var fs http.Handler
		// if isDevelopment {
		fs = http.FileServer(http.Dir("./assets"))
		// } else {
		// 	fs = http.FileServer(http.FS(assets.Assets))
		// }

		fs.ServeHTTP(w, r)
	})

	mux.Handle("GET /assets/", http.StripPrefix("/assets/", assetHandler))
}
