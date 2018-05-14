package core

import (
	"sync"
)


type Agregator struct {
	sync.Mutex
	exchangeBooks map[string]ExchangeBook
}

func NewAgregator() *Agregator {
	var agregator = Agregator{}
	agregator.exchangeBooks = map[string]ExchangeBook{}
	return &agregator
}

func (b *Agregator) add(exchangeBook ExchangeBook) {
	b.Lock()
	//fmt.Println("added:", exchangeBook)
	b.exchangeBooks[exchangeBook.Exchange.String()] = exchangeBook
	b.Unlock()
}

func (b *Agregator) getExchangeBooks()  []ExchangeBook {
	var tempBooks = []ExchangeBook{}
	b.Lock()
	for _,v := range b.exchangeBooks {
		tempBooks = append(tempBooks, v)
	}
	b.Unlock()

	return tempBooks
}