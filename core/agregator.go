package core

import (
	"sync"
)

type ExchangesBooks struct {
	mu            sync.Mutex
	exchangeBooks map[string]ExchangeBook
}


type Agregator struct {
	exchangeBooks ExchangesBooks
}

func NewAgregator() *Agregator {
	var agregator = Agregator{}
	agregator.exchangeBooks = ExchangesBooks{}
	agregator.exchangeBooks.exchangeBooks = map[string]ExchangeBook{"":newExchangeBook(Bitfinex)}
	delete(agregator.exchangeBooks.exchangeBooks, "")
	return &agregator
}

func (self *Agregator) add(exchangeBook ExchangeBook) {
	//fmt.Println("added:", exchangeBook)
	self.exchangeBooks.mu.Lock()
	self.exchangeBooks.exchangeBooks[exchangeBook.Exchange.String()] = exchangeBook
	self.exchangeBooks.mu.Unlock()
}

func (self *Agregator) getExchangeBooks()  map[string]ExchangeBook {

	self.exchangeBooks.mu.Lock()

	newExchangesBooks := map[string]ExchangeBook{"":newExchangeBook(Bitfinex)}
	delete(newExchangesBooks, "")

	for k,v := range  self.exchangeBooks.exchangeBooks {

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
	self.exchangeBooks.mu.Unlock()
	return newExchangesBooks
}