package core

import "time"

type CoinManager struct {
	BasicManager
	exchangeBook ExchangeBook
}


func (self *CoinManager) startSendingDataBack(exchangeConfiguration ExchangeConfiguration, resultChan chan Result) {

	for range time.Tick(1 * time.Second) {
		func() {


			mu.Lock()
			newBook := newExchangeBook(exchangeConfiguration.Exchange)

			for k,coinBook := range self.exchangeBook.CoinsBooks {
				newCoinBook := NewCoinBook(coinBook.Pair)

				for k,v := range coinBook.Asks {
					newCoinBook.Asks[k] = v
				}

				for k,v := range coinBook.Bids {
					newCoinBook.Bids[k] = v
				}

				newBook.CoinsBooks[k] = newCoinBook
			}
			mu.Unlock()

			resultChan <- Result{newBook, nil}
		}()
	}
}
