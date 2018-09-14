package core


import (
	//"strconv"
	//"github.com/btcsuite/btcutil"

	//"fmt"
	"fmt"
	//"github.com/ajph/bitstamp-go"
	"time"
	"log"
	//"debug/elf"
	"encoding/json"
	"github.com/ajph/bitstamp-go"
	"strconv"
)

const WS_TIMEOUT = 10 * time.Second


type BitstampManager struct {
	CoinManager
	api             *bitstamp.WebSocket
	//restApi *api.RestApi
}



func (self *BitstampManager) StartListen(exchangeConfiguration ExchangeConfiguration, getExchangeBookCompletion GetExchangeBookCompletion, tradeCompletion WsTradeCompletion) {

	//self.restApi = api.NewRestApi()
	self.exchangeBook = newExchangeBook(Bitstamp)

	go self.getApiOrderBook()
	go self.startSendingDataBack(exchangeConfiguration, getExchangeBookCompletion)



	for {
		log.Println("Dialing Bitstamp...")
		var err error
		self.api, err = bitstamp.NewWebSocket(WS_TIMEOUT)
		if err != nil {
			log.Printf("Error connecting: %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
		self.api.Subscribe("diff_order_book")
		self.api.Subscribe("live_trades")


		//restApiResponseChan := make(chan api.RestApiReposponse)

		//urlString := "https://api.bitfinex.com/v2/book/tBTCUSD/P2?len=100"
		//go self.restApi.PublicRequest(urlString, restApiResponseChan)



	L:
		for {
			select {
			case ev := <-self.api.Stream:
				self.handleEvent(ev, self.api, tradeCompletion)

			case err := <-self.api.Errors:
				log.Printf("Bitstamp Socket error: %s, reconnecting...", err)
				self.api.Close()
				break L

			case <-time.After(10 * time.Second):
				self.api.Ping()

			}
		}
	}
}




func (self *BitstampManager) handleEvent(e *bitstamp.Event, Ws *bitstamp.WebSocket, tradeCompletion WsTradeCompletion) {
	switch e.Event {
	// pusher stuff
	case "pusher:connection_established":
		log.Println("Bitstamp Connected")
	case "pusher_internal:subscription_succeeded":
		log.Println("Bitstamp Subscribed for:", e.Channel)
	case "pusher:pong":
		// ignore
	case "pusher:ping":
		Ws.Pong()

		// bitstamp
	case "trade":
		event := BitstampTrade{}
		json.Unmarshal([]byte(e.Data.(string)), &event)
		self.handleTrade(&event, tradeCompletion)
		//log.Println(event)
	case "data":
		//log.Println(e.Data)
		orderBookResult := bitstamp.OrderBookResult{}
		json.Unmarshal([]byte(e.Data.(string)), &orderBookResult)





		//fmt.Println(orderBookResult.Asks)
		self.addEvent(orderBookResult)
		// other
	default:
		log.Printf("Unknown event: %#v\n", e)
	}
}


func (self *BitstampManager) getApiOrderBook() {
	apiOrderBook, _ := bitstamp.OrderBook("btcusd")

	//fmt.Println(apiOrderBook.Bids)
	self.addEvent(*apiOrderBook)
}



func (self *BitstampManager) addEvent(orderBookResult bitstamp.OrderBookResult)  {

	mu.Lock()

	symbol := "BTC/USD"

	if _, ok := self.exchangeBook.CoinsBooks[symbol]; !ok {
		newCoinBook := NewCoinBook(CurrencyPair{BritCoin, Tether})
		self.exchangeBook.CoinsBooks[symbol] = newCoinBook
	}


	previouseCoinBook := self.exchangeBook.CoinsBooks[symbol]


	for _, ask := range orderBookResult.Asks {

		if ask.Amount == 0 {
			delete(previouseCoinBook.Asks, ask.Price)
		} else if ask.Amount > 0 {
			previouseCoinBook.Asks[ask.Price] = ask.Amount
		} else {
			fmt.Println("amount can't be:", ask.Amount)
		}
	}


	for _, bid := range orderBookResult.Bids {

		if bid.Amount == 0 {
			delete(previouseCoinBook.Bids, bid.Price)
		} else if bid.Amount > 0 {
			previouseCoinBook.Bids[bid.Price] = bid.Amount
		} else {
			fmt.Println("amount can't be:", bid.Amount)
		}
	}

	self.exchangeBook.CoinsBooks[symbol] = previouseCoinBook
	mu.Unlock()
}


func (self *BitstampManager) handleTrade(event *BitstampTrade, tradeCompletion WsTradeCompletion)  {
	trade := WsTrade{}
	trade.Exchange = Bitstamp.String()
	trade.Symbol = "BTC/USD"
	trade.Quantity = event.Amount
	trade.Price = event.Price
	time, _ := strconv.ParseInt(event.Timestamp, 0, 64)
	trade.TradeTime = time * 1000 // to get from 1536617724 -> 1536617724000
	if event.Type == 0 {
		trade.IsBid = true
	} else {
		trade.IsBid = false
	}
	tradeCompletion(&trade)

}

type BitstampTrade struct {
	Amount      float64 `json:"amount"`
	BuyOrderID  int     `json:"buy_order_id"`
	SellOrderID int     `json:"sell_order_id"`
	AmountStr   string  `json:"amount_str"`
	PriceStr    string  `json:"price_str"`
	Timestamp   string  `json:"timestamp"`
	Price       float64 `json:"price"`
	Type        int     `json:"type"`
	ID          int     `json:"id"`
}