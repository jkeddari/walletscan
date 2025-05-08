package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port           int
	ethereumClient *ethclient.Client
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	eClient, err := ethclient.Dial("https://mainnet.infura.io/v3/4a8a21bb79a941559477173d40a3901b")
	if err != nil {
		log.Fatal(err)
	}

	routes := &Server{
		port:           port,
		ethereumClient: eClient,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", routes.port),
		Handler:      routes.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
