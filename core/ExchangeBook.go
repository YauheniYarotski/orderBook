package core


type ExchangeBook struct {
	ExchangeTitle string `json:"exchange_title"`
	Exchange Exchange  `json:"exchange"`
	CoinsBooks map[string]CoinBook  `json:"books"`
}

type WSExchangeBook struct {
	ExchangeTitle string `json:"exchange_title"`
	CoinsBooks []WsCoinBook  `json:"books"`
}

func newExchangeBook(exchange Exchange) ExchangeBook  {
	exchangeBook := ExchangeBook{}

	exchangeBook.Exchange = exchange
	exchangeBook.CoinsBooks = map[string]CoinBook{"":NewCoinBook(CurrencyPair{})}
	delete(exchangeBook.CoinsBooks, "")
	return exchangeBook
}
