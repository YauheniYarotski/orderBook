package core

import (
	"sync"
	"fmt"
)


type Agregator struct {
	sync.Mutex
	exchangeEvents ExchangeEvents
}

func NewAgregator() *Agregator {
	var agregator = Agregator{}
	agregator.exchangeEvents = ExchangeEvents{}
	return &agregator
}

func (b *Agregator) add(exchangeEvents ExchangeEvents) {
	b.Lock()
	fmt.Println("added:", exchangeEvents)
	for k,v := range exchangeEvents {
		b.exchangeEvents[k] = v
	}
	b.Unlock()
}

//func (b *Agregator) getEvents() map[string]Events {
//	b.Lock()
//	var tempEvents = map[string]Events{}
//	for k,v := range b.events {
//		tempEvents[k] = v
//	}
//	b.Unlock()
//	return tempEvents
//}
