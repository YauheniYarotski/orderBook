package core

import (
	"sync"
)





type Agregator struct {
	mu            *sync.Mutex
	exchangeBooks map[string]ExchangeBook
}

func NewAgregator() Agregator {
	var agregator = Agregator{}
	agregator.mu =  &sync.Mutex{}
	agregator.exchangeBooks = map[string]ExchangeBook{"":newExchangeBook(Bitfinex)}
	delete(agregator.exchangeBooks, "")
	return agregator
}

func (self *Agregator) add(exchangeBook ExchangeBook) {
	//fmt.Println("added:", exchangeBook)
	self.mu.Lock()
	self.exchangeBooks[exchangeBook.Exchange.String()] = exchangeBook
	self.mu.Unlock()
}

func (self *Agregator) getExchangeBooks()  map[string]ExchangeBook {

	self.mu.Lock()

	newExchangesBooks := map[string]ExchangeBook{"":newExchangeBook(Bitfinex)}
	delete(newExchangesBooks, "")

	for k,v := range  self.exchangeBooks {

		newBook := newExchangeBook(v.Exchange)

		for k,coinBook := range v.CoinsBooks {
			newCoinBook := NewCoinBook(coinBook.Pair)

			for k,v := range coinBook.Asks {
				newCoinBook.Asks[k] = v
			}

			for k,v := range coinBook.Bids {
				newCoinBook.Bids[k] = v
			}

			newBook.CoinsBooks[k] = newCoinBook
		}

		newExchangesBooks[k] = newBook

	}
	self.mu.Unlock()
	return newExchangesBooks
}