package core


type ExchangeBook struct {
	ExchangeTitle string `json:"exchange_title"`
	Exchange Exchange  `json:"exchange"`
	CoinsBooks map[string]CoinBook  `json:"books"`
}