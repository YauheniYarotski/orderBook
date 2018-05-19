package core

import (
	"sync"
)

var Lock sync.Mutex

type Agregator struct {
	exchangeBooks map[string]ExchangeBook
}

func NewAgregator() *Agregator {
	var agregator = Agregator{}
	agregator.exchangeBooks = map[string]ExchangeBook{"":newExchangeBook(Bitfinex)}
	delete(agregator.exchangeBooks, "")
	return &agregator
}

func (b *Agregator) add(exchangeBook ExchangeBook) {
	//fmt.Println("added:", exchangeBook)
	Lock.Lock()
	b.exchangeBooks[exchangeBook.Exchange.String()] = exchangeBook
	Lock.Unlock()
}

func (b *Agregator) getExchangeBooks()  map[string]ExchangeBook {
	Lock.Lock()
	tmp := b.exchangeBooks
	Lock.Unlock()
	return tmp
}