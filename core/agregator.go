package core

import (
	"sync"
)
var mu = &sync.Mutex{}

type Agregator struct {

	exchangeBooks map[string]ExchangeBook
	trades []*WsTrade
	booksCh chan *ExchangeBook
	signalCh chan bool
	getbooksCh chan map[string]ExchangeBook

}

func NewAgregator() *Agregator {
	var agregator = Agregator{}
	agregator.exchangeBooks = map[string]ExchangeBook{"":newExchangeBook(Bitfinex)}
	delete(agregator.exchangeBooks, "")
	agregator.trades = []*WsTrade{}
	agregator.booksCh = make(chan *ExchangeBook)

	agregator.signalCh = make(chan bool)


	agregator.getbooksCh = make(chan map[string]ExchangeBook)
	go agregator.startListen()
	return &agregator
}

func (self *Agregator) startListen() {
	for {
		select {
		// send message to the client
		case exchangeBook := <-self.booksCh:
			self.exchangeBooks[exchangeBook.Exchange.String()] = *exchangeBook
		case <-self.signalCh:
			self.getbooksCh <- self.exchangeBooks
			//self.exchangeBooks[exchangeBook.Exchange.String()] = *exchangeBook
		}
	}
}

func (self *Agregator) add(exchangeBook *ExchangeBook) {
	//fmt.Println("added:", exchangeBook)
	self.booksCh <- exchangeBook
	//mu.Lock()
	//self.exchangeBooks[exchangeBook.Exchange.String()] = *exchangeBook
	//mu.Unlock()
}

func (self *Agregator) getExchangeBooks(granulation float64)  map[string]ExchangeBook {


	newExchangesBooks := map[string]ExchangeBook{"":newExchangeBook(Bitfinex)}
	delete(newExchangesBooks, "")
	self.signalCh <- true
	exchangeBooks := <- self.getbooksCh
	for k,v := range  exchangeBooks {

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
	return newExchangesBooks
}


func (self *Agregator) addTrade(trade *WsTrade) {
	self.trades = append(self.trades, trade)
	if len(self.trades) > 1000 {
		self.trades = self.trades[100: len(self.trades)-1]
	}
}