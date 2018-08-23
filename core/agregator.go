package core

import (
	"sync"
	"github.com/adshao/go-binance"
)
var mu = &sync.Mutex{}

type Agregator struct {

	exchangeBooks map[string]ExchangeBook
	trades []*binance.WsTradeEvent
}

func NewAgregator() *Agregator {
	var agregator = Agregator{}
	agregator.exchangeBooks = map[string]ExchangeBook{"":newExchangeBook(Bitfinex)}
	delete(agregator.exchangeBooks, "")
	agregator.trades = []*binance.WsTradeEvent{}
	return &agregator
}

func (self *Agregator) add(exchangeBook ExchangeBook) {
	//fmt.Println("added:", exchangeBook)
	mu.Lock()
	self.exchangeBooks[exchangeBook.Exchange.String()] = exchangeBook
	mu.Unlock()
}

func (self *Agregator) getExchangeBooks(granulation float64)  map[string]ExchangeBook {

	mu.Lock()

	newExchangesBooks := map[string]ExchangeBook{"":newExchangeBook(Bitfinex)}
	delete(newExchangesBooks, "")

	for k,v := range  self.exchangeBooks {

		newBook := newExchangeBook(v.Exchange)
		newBook.ExchangeTitle = v.Exchange.String()

		for k,coinBook := range v.CoinsBooks {
			newCoinBook := NewCoinBook(coinBook.Pair)

			for k,f := range coinBook.Asks {
				k = Trunc(k, granulation)
				newCoinBook.Asks[k] = newCoinBook.Asks[k] + f
			}

			for k,f := range coinBook.Bids {
				k = Trunc(k, granulation)
				newCoinBook.Bids[k] = newCoinBook.Bids[k] + f

			}

			newBook.CoinsBooks[k] = newCoinBook
		}

		newExchangesBooks[k] = newBook

	}
	mu.Unlock()
	return newExchangesBooks
}


func (self *Agregator) addTrade(trade *binance.WsTradeEvent) {
	self.trades = append(self.trades, trade)
	if len(self.trades) > 1000 {
		self.trades = self.trades[100: len(self.trades)-1]
	}
}