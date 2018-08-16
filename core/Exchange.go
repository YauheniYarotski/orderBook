package core

import "strings"

type Exchange int

func NewExchange(exchangeString string) Exchange {
	exchanges := map[string]Exchange{"BINANCE": Binance, "BITFINEX": Bitfinex, "GDAX": Gdax, "HITBTC": HitBtc, "OKEX": Okex, "POLONIEX": Poloniex, "BITTREX": Bittrex, "HUOBI": Huobi, "UPBIT": Upbit, "KRAKEN": Kraken, "BITHUMB": Bithumb, "BITMEX": Bitmex}
	exchange := exchanges[strings.ToUpper(exchangeString)]
	return exchange
}

func (exchange Exchange) String() string {
	exchanges := [...]string{
		"BINANCE",
		"BITFINEX",
		"GDAX",
		"HITBTC",
		"OKEX",
		"POLONIEX",
		"BITTREX",
		"HUOBI",
		"UPBIT",
		"KRAKEN",
		"BITHUMB",
		"BITMEX"}
	return exchanges[exchange]
}

const (
	Binance  Exchange = 0
	Bitfinex Exchange = 1
	Gdax     Exchange = 2
	HitBtc   Exchange = 3
	Okex     Exchange = 4
	Poloniex Exchange = 5
	Bittrex  Exchange = 6
	Huobi 	 Exchange = 7
	Upbit 	 Exchange = 8
	Kraken 	 Exchange = 9
	Bithumb  Exchange = 10
	Bitmex   Exchange = 11
)

type ExchangeConfiguration struct {
	Exchange            Exchange
	RefreshInterval     int
	Pairs []CurrencyPair
}