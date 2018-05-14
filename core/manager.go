package core

import (
"strings"
"sync"
"time"

	"github.com/KristinaEtc/slf"
)

var log = slf.WithContext("core")

const maxTickerAge = 5

type BasicManager struct {
	sync.Mutex
	//tickers map[string]Ticker
}

type PriceLevels struct {
	Asks map[string]string
	Bids map[string]string
}

type CoinEvents map[string]PriceLevels

type ExchangeEvents map[string]CoinEvents

type Result struct {
	ExchangeEvents ExchangeEvents
	Err              *error
}

type Manager struct {
	binanceManager  *BinanceManager
	//hitBtcManager   *HitBtcManager
	//poloniexManager *PoloniexManager
	bitfinexManager *BitfinexManager
	//gdaxManager     *GdaxManager
	//okexManager     *OkexManager
	//bittrexManager     *BittrexManager
	//huobiManager     *HuobiManager
	//upbitManager     *UpbitManager
	//krakenManager     *KrakenManager
	//bithumbManager     *BithumbManager


	agregator *Agregator
}

func NewManager() *Manager {
	var manger = Manager{}

	manger.agregator = NewAgregator()
	manger.binanceManager = NewBinanceManager()
	//manger.hitBtcManager = &HitBtcManager{}
	//manger.poloniexManager = &PoloniexManager{}
	manger.bitfinexManager = &BitfinexManager{}
	//manger.gdaxManager = &GdaxManager{}
	//manger.okexManager = &OkexManager{}
	//manger.server = &stream.Server{}
	//manger.bittrexManager = &BittrexManager{}
	//manger.huobiManager = &HuobiManager{}
	//manger.upbitManager = &UpbitManager{}
	//manger.krakenManager = &KrakenManager{}
	//manger.bithumbManager = &BithumbManager{}

	return &manger
}

type ManagerConfiguration struct {
	TargetCurrencies    []string        `json:"targetCurrencies"`
	ReferenceCurrencies []string        `json:"referenceCurrencies"`
	Exchanges           []string        `json:"exchanges"`
	RefreshInterval     time.Duration   `json:"refreshInterval"`
	DBConfiguration     DBConfiguration `json:"dbconfiguration"`
}

func (b *ManagerConfiguration) Pairs() []CurrencyPair {
	var pairs = []CurrencyPair{}
	for _, targetCurrency := range b.TargetCurrencies {
		for _, referenceCurrency := range b.ReferenceCurrencies {

			if referenceCurrency == "USD" {
				referenceCurrency = "USDT"
			} else if referenceCurrency == targetCurrency {
				continue
			}
			pair := CurrencyPair{NewCurrencyWithCode(targetCurrency), NewCurrencyWithCode(referenceCurrency)}
			pairs = append(pairs, pair)
		}
	}
	return pairs
}

type DBConfiguration struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type Exchange int

func NewExchange(exchangeString string) Exchange {
	exchanges := map[string]Exchange{"BINANCE": Binance, "BITFINEX": Bitfinex, "GDAX": Gdax, "HITBTC": HitBtc, "OKEX": Okex, "POLONIEX": Poloniex, "BITTREX": Bittrex, "HUOBI": Huobi, "UPBIT": Upbit, "KRAKEN": Kraken, "BITHUMB": Bithumb}
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
		"BITHUMB"}
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
	Bithumb Exchange = 10
)

type ExchangeConfiguration struct {
	Exchange            Exchange
	TargetCurrencies    []string
	ReferenceCurrencies []string
	RefreshInterval     int
	Pairs []CurrencyPair
}

func (b *Manager) launchExchange(exchangeConfiguration ExchangeConfiguration, ch chan Result) {

	switch exchangeConfiguration.Exchange {
	case Binance:
		go b.binanceManager.StartListen(exchangeConfiguration, ch)
	case Bitfinex:
		go b.bitfinexManager.StartListen(exchangeConfiguration, ch)
	//case Gdax:
	//	go b.gdaxManager.StartListen(exchangeConfiguration, ch)
	//case HitBtc:
	//	go b.hitBtcManager.StartListen(exchangeConfiguration, ch)
	//case Okex:
	//	go b.okexManager.StartListen(exchangeConfiguration, ch)
	//case Poloniex:
	//	go b.poloniexManager.StartListen(exchangeConfiguration, ch)
	//case Bittrex:
	//	go b.bittrexManager.StartListen(exchangeConfiguration, ch)
	//case Huobi:
	//	go b.huobiManager.StartListen(exchangeConfiguration, ch)
	//case Upbit:
	//	go b.upbitManager.StartListen(exchangeConfiguration, ch)
	//case Kraken:
	//	go b.krakenManager.StartListen(exchangeConfiguration, ch)
	//case Bithumb:
	//	go b.bithumbManager.StartListen(exchangeConfiguration, ch)
	default:
		log.Errorf("launchExchange:default %v", exchangeConfiguration.Exchange.String())
	}
}

func (b *Manager) StartListen(configuration ManagerConfiguration) {

	ch := make(chan Result)

	for _, exchangeString := range configuration.Exchanges {
		exchangeConfiguration := ExchangeConfiguration{}
		exchangeConfiguration.Exchange = NewExchange(exchangeString)
		exchangeConfiguration.TargetCurrencies = configuration.TargetCurrencies
		exchangeConfiguration.ReferenceCurrencies = configuration.ReferenceCurrencies
		exchangeConfiguration.Pairs = configuration.Pairs()
		b.launchExchange(exchangeConfiguration, ch)
	}



	for {
		select {
		case result := <-ch:

			if result.Err != nil {
				log.Errorf("StartListen:error: %v", result.Err)
			} else {
				//fmt.Println(result.ExchangeEvents)
				//b.agregator.add(*result.TickerCollection, result.exchangeTitle)
				b.agregator.add(result.ExchangeEvents)
			}

		}
	}

}
//
//func (b *Manager) convertToTickerCollection(tickerCollection TickerCollection) stream.StreamTickerCollection {
//	var streamTickerCollection = stream.StreamTickerCollection{}
//	var streamTickers = []stream.StreamTicker{}
//
//	streamTickerCollection.TimpeStamp = tickerCollection.TimpeStamp
//	for _, ticker := range tickerCollection.Tickers {
//		var streamTicker = b.convertToStreamTicker(ticker)
//		streamTickers = append(streamTickers, streamTicker)
//	}
//	streamTickerCollection.Tickers = streamTickers
//
//	return streamTickerCollection
//
//}
//
//func (b *Manager) convertToStreamTicker(ticker Ticker) stream.StreamTicker {
//	var streamTicker = stream.StreamTicker{}
//	streamTicker.Rate = ticker.Rate
//	streamTicker.Pair = ticker.Pair
//	return streamTicker
//}
