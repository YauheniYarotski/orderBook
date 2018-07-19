package core

import "sync"

type ExchangeBook struct {
	mu      sync.Mutex
	Exchange Exchange  `json:"exchange"`
	CoinsBooks map[string]CoinBook  `json:"books"`
}