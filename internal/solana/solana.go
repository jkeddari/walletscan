package solana

import (
	"context"
	"encoding/json"
	"math"
	"strconv"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/jkeddari/walletscan/internal/gecko"
	"github.com/jkeddari/walletscan/internal/types"
)

type TokenBalance struct {
	MintAddress string
	Amount      float64
	Decimals    uint8
}

func ValidAddress(address string) bool {
	if _, err := solana.PublicKeyFromBase58(address); err != nil {
		return false
	}
	return true
}

func Scan(address string, tokenDetails *gecko.CoinsDetails, prices gecko.Prices) (wallet *types.WalletData, err error) {
	result := types.NewWalletData()

	client := rpc.New(rpc.MainNetBeta_RPC)

	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return nil, err
	}

	balanceResp, err := client.GetBalance(context.Background(), pubKey, rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}
	solBalance := float64(balanceResp.Value) / float64(solana.LAMPORTS_PER_SOL)

	result.Add("solana", "solana", "SOL", "Solana", tokenDetails.Icon("solana"), solBalance, prices.PriceByID("solana"))

	tokenAccountsResp, err := client.GetTokenAccountsByOwner(
		context.Background(),
		pubKey,
		&rpc.GetTokenAccountsConfig{
			ProgramId: solana.TokenProgramID.ToPointer(),
		},
		&rpc.GetTokenAccountsOpts{Encoding: solana.EncodingJSONParsed},
	)
	if err != nil {
		return nil, err
	}

	for _, acc := range tokenAccountsResp.Value {
		data := acc.Account.Data.GetRawJSON()

		var parsed struct {
			Parsed struct {
				Info struct {
					Mint        string `json:"mint"`
					TokenAmount struct {
						Amount   string `json:"amount"`
						Decimals uint8  `json:"decimals"`
					} `json:"tokenAmount"`
				} `json:"info"`
			} `json:"parsed"`
		}

		if err := json.Unmarshal(data, &parsed); err != nil {
			continue
		}

		amountInt, err := strconv.ParseFloat(parsed.Parsed.Info.TokenAmount.Amount, 64)
		if err != nil {
			continue
		}
		decimals := parsed.Parsed.Info.TokenAmount.Decimals
		mint := parsed.Parsed.Info.Mint

		amountFloat := float64(amountInt) / math.Pow10(int(decimals))

		if tokenID, tokenInfo := tokenDetails.CoinByContract("solana", mint); tokenID != "" && tokenInfo != nil {
			result.Add(
				tokenID,
				"solana",
				tokenInfo.Symbol,
				tokenInfo.Name,
				tokenInfo.Icon,
				amountFloat,
				prices.PriceByID(tokenID),
			)
		}
	}

	return result, nil
}
