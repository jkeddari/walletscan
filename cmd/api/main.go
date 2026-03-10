package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/a-h/templ"
	"github.com/jkeddari/walletscan"
	"github.com/jkeddari/walletscan/internal/gecko"
	"github.com/jkeddari/walletscan/internal/types"
	"github.com/jkeddari/walletscan/ui/modules"
	"github.com/jkeddari/walletscan/ui/pages"
	"github.com/joho/godotenv"
)

const (
	defaultPort = "8090"
)

var (
	coinsInfo       *gecko.CoinsDetails
	coinGeckoAPIKey string
	logger          = slog.New(slog.NewJSONHandler(os.Stdout, nil))
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
	setupSEORoutes(mux)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("new connection", "address", r.RemoteAddr)
		pages.Landing().Render(context.Background(), w)
	})

	mux.Handle("GET /about", templ.Handler(pages.About()))
	mux.HandleFunc("POST /scan-address", func(w http.ResponseWriter, r *http.Request) {
		address := r.FormValue("address")
		if address == "" {
			http.Error(w, "Address required", http.StatusBadRequest)
			return
		}

		addressType := walletscan.IsAddress(address)
		if addressType == types.UnknownAddress {
			logger.Error("bad address", "address", address)
			w.WriteHeader(http.StatusOK)
			modules.ScanForm(address, true).Render(r.Context(), w)
			return
		}

		logger.Info("scan address", "type", addressType.String(), "address", address)

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

		sortKey := r.URL.Query().Get("sort")
		if sortKey == "" {
			sortKey = "symbol"
		}
		order := r.URL.Query().Get("order")
		asc := order != "desc"

		data, err := walletscan.Scan(address, coinsInfo, prices)
		if err != nil {
			logger.Error("scan error", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		pages.Scan(address, data, sortKey, asc).Render(context.Background(), w)
	})

	logger.Info("Server is running", "port", port, "coins", len(coinsInfo.IDs))
	http.ListenAndServe(":"+port, mux)
}

func setupAssetsRoutes(mux *http.ServeMux) {
	assetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var fs http.Handler
		fs = http.FileServer(http.Dir("./assets"))
		fs.ServeHTTP(w, r)
	})

	mux.Handle("GET /assets/", http.StripPrefix("/assets/", assetHandler))
}

func setupSEORoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Clean("assets/robots.txt"))
	})
	mux.HandleFunc("GET /sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Clean("assets/sitemap.xml"))
	})
}
