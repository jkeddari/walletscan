package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"

	"github.com/jkeddari/walletscan/internal/web"

	"github.com/a-h/templ"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Web API
	fileServer := http.FileServer(http.FS(web.AssetsFiles))
	r.Handle("/assets/*", fileServer)
	r.Get("/", templ.Handler(web.BalanceForm()).ServeHTTP)
	r.Post("/balance", s.balanceWebHandler)

	// Public API
	r.Get("/balance/{address}", s.ethereumBalanceHandler)

	return r
}

func weiToEther(wei *big.Int) float64 {
	bfloat, _ := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether)).Float64()
	return bfloat
}

func (s *Server) balanceWebHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	// TODO handle error
	amount, _ := s.ethereumBalance(r.FormValue("address"))
	component := web.BalancePost(amount)

	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error rendering in HelloWebHandler: %e", err)
	}
}

func (s *Server) ethereumBalanceHandler(w http.ResponseWriter, r *http.Request) {
	// TODO handle error
	amount, _ := s.ethereumBalance(chi.URLParam(r, "address"))
	resp := map[string]string{
		"amount": amount,
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}
}

func (s *Server) ethereumBalance(rawAddress string) (string, error) {
	ctx := context.Background()

	if !common.IsHexAddress(rawAddress) {
		return "", errors.New("bad address")
	}

	address := common.HexToAddress(rawAddress)

	block, err := s.ethereumClient.BlockNumber(ctx)
	if err != nil {
		return "", err
	}

	amount, err := s.ethereumClient.BalanceAt(ctx, address, big.NewInt(int64(block)))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%f", weiToEther(amount)), nil
}
