package core

import (
"time"
)

const maxTickerAge = 5

type BasicManager struct {
	//tickers map[string]Ticker
}




type Result struct {
	ExchangeBook ExchangeBook
	Err              *error
}


type Manager struct {
	binanceManager  *BinanceManager
	//hitBtcManager   *HitBtcManager
	//poloniexManager *PoloniexManager
	bitfinexManager *BitfinexManager
	bitmexManager *BitmexManager
	bitstampManager *BitstampManager
	dbManger            *DbManager
	wsServer *WsServer
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
	manger.bitmexManager = &BitmexManager{}
	manger.bitstampManager = &BitstampManager{}
	manger.wsServer = NewWsServer("/books")
	//manger.gdaxManager = &GdaxManager{}
	//manger.okexManager = &OkexManager{}
	//manger.server = &stream.Server{}
	//manger.bittrexManager = &BittrexManager{}
	//manger.huobiManager = &HuobiManager{}
	//manger.upbitManager = &UpbitManager{}
	//manger.krakenManager = &KrakenManager{}
	//manger.bithumbManager = &BithumbManager{}

	dbConfig := DBConfiguration{}
	dbConfig.Name = "test"
	dbConfig.Password = "postgres"
	dbConfig.User = "postgres"

	manger.dbManger = NewDbManager(dbConfig)
	return &manger
}

type ManagerConfiguration struct {
	Exchanges           []string        `json:"exchanges"`
	RefreshInterval     time.Duration   `json:"refreshInterval"`
	DBConfiguration     DBConfiguration `json:"dbconfiguration"`
}




func (b *Manager) launchExchange(exchangeConfiguration ExchangeConfiguration, ch chan Result, tradeCh chan *WsTrade) {

	switch exchangeConfiguration.Exchange {
	case Binance:
		go b.binanceManager.StartListen(exchangeConfiguration, ch, tradeCh)
	case Bitfinex:
		go b.bitfinexManager.StartListen(exchangeConfiguration, ch, tradeCh)
	case Bitmex:
		go b.bitmexManager.StartListen(exchangeConfiguration, ch)
	case Bitstamp:
		go b.bitstampManager.StartListen(exchangeConfiguration, ch)
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
		//log.Errorf("launchExchange:default %v", exchangeConfiguration.Exchange.String())
	}
}

func (self *Manager) Start(configuration ManagerConfiguration) {

	//start ws service
	go self.wsServer.start()
	self.wsServer.ServerHandler = func(granulation float64 ,exchangeBooks *[]ExchangeBook) {
		v := self.agregator.getExchangeBooks(granulation)

		for _,vv := range  v  {

			*exchangeBooks = append(*exchangeBooks, vv)
		}

		//tmp := v
		//*exchangeBooks = tmp

	}

	//go b.fillDb()


	ch := make(chan Result)
	tradeCh := make(chan *WsTrade)

	for _, exchangeString := range configuration.Exchanges {
		exchangeConfiguration := ExchangeConfiguration{}
		exchangeConfiguration.Exchange = NewExchange(exchangeString)
		self.launchExchange(exchangeConfiguration, ch, tradeCh)
	}


	for {
		select {
		case result := <-ch:

			if result.Err != nil {
				//log.Errorf("StartListen:error: %v", result.Err)
			} else {
				//fmt.Println(result.ExchangeEvents)
				//b.agregator.add(*result.TickerCollection, result.exchangeTitle)
				self.agregator.add(result.ExchangeBook)
			}

		case trade := <-tradeCh:
			self.agregator.addTrade(trade)
			self.sendTradeToWs(trade)
		}
	}
}

func (self *Manager)sendTradeToWs(trade *WsTrade) {
	if trade.Quantity >= 0.5 {
		self.wsServer.SendTrade(trade)
	}
}

//func (b *Manager) fillDb() {
//
//	for range time.Tick(1 * time.Second) {
//			b.dbManger.FillDb(b.agregator.getExchangeBooks())
//
//
//		//v := b.GetRates(time.Now().Add(-4 * time.Minute), "BINANCE", "BTS", []string{"BTC", "USDT"})
//		//
//		//for _,value := range v {
//		//	fmt.Println("get rates :", value.symbol(), value.TimpeStamp, value.Rate)
//		//}
//	}
//}

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


