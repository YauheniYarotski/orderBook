package core

import (
	"sync"
)
var mu = &sync.Mutex{}

type Agregator struct {

	exchangeBooks map[string]ExchangeBook
}

func NewAgregator() *Agregator {
	var agregator = Agregator{}
	agregator.exchangeBooks = map[string]ExchangeBook{"":newExchangeBook(Bitfinex)}
	delete(agregator.exchangeBooks, "")
	return &agregator
}

func (self *Agregator) add(exchangeBook ExchangeBook) {
	//fmt.Println("added:", exchangeBook)
	mu.Lock()
	self.exchangeBooks[exchangeBook.Exchange.String()] = exchangeBook
	mu.Unlock()
}

func (self *Agregator) getExchangeBooks()  map[string]ExchangeBook {

	mu.Lock()

	newExchangesBooks := map[string]ExchangeBook{"":newExchangeBook(Bitfinex)}
	delete(newExchangesBooks, "")

	for k,v := range  self.exchangeBooks {

		newBook := newExchangeBook(v.Exchange)
		newBook.ExchangeTitle = v.Exchange.String()

		for k,coinBook := range v.CoinsBooks {
			newCoinBook := NewCoinBook(coinBook.Pair)

			for k,f := range coinBook.Asks {

				if f >= 1.0 {
					newCoinBook.Asks[k] = f
				}
			}

			for k,f := range coinBook.Bids {
				if f >= 1.0 {
					newCoinBook.Bids[k] = f
				}
			}

			newBook.CoinsBooks[k] = newCoinBook
		}

		newExchangesBooks[k] = newBook

	}
	mu.Unlock()
	return newExchangesBooks
}