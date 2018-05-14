package core

import (
	"sync"
	"fmt"
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
	fmt.Println("added:", exchangeBook)
	b.exchangeBooks[exchangeBook.Exchange.String()] = exchangeBook
	b.Unlock()
}