package types

import (
	"fmt"
	"strings"
	"sync"
)

type AddressType int

const (
	UnknownAddress = iota
	EVMAddress
	BitcoinAddress
	SolanaAddress
)

// Balance defines a total balance for a token ID.
type Balance struct {
	TokenID string

	// Amount defines token amount by network.
	Amount map[string]float64

	// Symbol defines token symbol.
	Symbol string

	// Name define token name.
	Name string

	// Price defines token price.
	Price float64

	// TotalAmount defines token total amount.
	TotalAmount float64

	// Expanded
	Expanded bool
}

func NewBalance(id, network, symbol, name string, amount, price float64) *Balance {
	return &Balance{
		TokenID: id,
		Amount: map[string]float64{
			network: amount,
		},
		Symbol:      symbol,
		Name:        name,
		Price:       price,
		TotalAmount: amount,
	}
}

func (b *Balance) add(network string, amount float64) {
	if b.Amount == nil {
		b.Amount = map[string]float64{
			network: amount,
		}
		b.TotalAmount = amount
		return
	}

	if _, ok := b.Amount[network]; ok {
		b.Amount[network] += amount
		b.TotalAmount += amount
		return
	}

	b.Amount[network] = amount
	b.TotalAmount = amount
}

func (b *Balance) Value() float64 {
	return b.Price * b.TotalAmount
}

// WalletData defines all wallet's balance.
type WalletData struct {
	// TotalValue defines wallet total value in USD.
	TotalValue float64

	// Network defines networks where token are contained.
	Network map[string]struct{}

	// Balance defines balance by token ID.
	Balances map[string]*Balance

	m sync.Mutex
}

func NewWalletData() *WalletData {
	return &WalletData{
		Network:  make(map[string]struct{}),
		Balances: make(map[string]*Balance),
	}
}

func (w *WalletData) TokenNumber() int {
	return len(w.Balances)
}

func (w *WalletData) Networks() []string {
	var networks []string
	for n := range w.Network {
		networks = append(networks, n)
	}
	return networks
}

func (w *WalletData) Add(id, network, symbol, name string, amount, price float64) {
	if amount == 0 {
		return
	}

	w.m.Lock()
	defer w.m.Unlock()

	w.Network[network] = struct{}{}

	w.TotalValue += amount * price
	if w.Balances[id] == nil {
		w.Balances[id] = NewBalance(id, network, symbol, name, amount, price)
		return
	}

	w.Balances[id].add(network, amount)
}

func (w *WalletData) Print() {
	w.m.Lock()
	defer w.m.Unlock()

	fmt.Println("======= Wallet Data =======")
	fmt.Printf("Total Value: $%.2f\n", w.TotalValue)
	fmt.Println("Networks:", w.formatNetworks())
	fmt.Println("\nBalances:")

	for tokenID, balance := range w.Balances {
		fmt.Printf("\nToken ID     : %s\n", tokenID)
		fmt.Printf("  Name        : %s\n", balance.Name)
		fmt.Printf("  Symbol      : %s\n", balance.Symbol)
		fmt.Printf("  Price       : $%.4f\n", balance.Price)
		fmt.Printf("  TotalAmount : %.4f\n", balance.TotalAmount)
		fmt.Println("  Amount by Network:")
		for network, amount := range balance.Amount {
			fmt.Printf("    - %s: %.4f\n", network, amount)
		}
	}
	fmt.Printf("===========================\n\n")
}

func (w *WalletData) formatNetworks() string {
	var keys []string
	for k := range w.Network {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}
