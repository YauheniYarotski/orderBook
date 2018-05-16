package core

import (
	"sync"
)


type Agregator struct {
	exchangeBooks sync.Map //map[string]ExchangeBook
}

func NewAgregator() *Agregator {
	var agregator = Agregator{}
	agregator.exchangeBooks = sync.Map{}
	return &agregator
}

func (b *Agregator) add(exchangeBook ExchangeBook) {
	//fmt.Println("added:", exchangeBook)
	b.exchangeBooks.Store(exchangeBook.Exchange.String(), exchangeBook)
}

func (b *Agregator) getExchangeBooks()  []ExchangeBook {
	var tempBooks = []ExchangeBook{}
	b.exchangeBooks.Range(func(k, v interface{}) bool {
		tempBooks = append(tempBooks, v.(ExchangeBook))
		return true
	})
	return tempBooks
}