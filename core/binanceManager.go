
package core

import (
"encoding/json"

	"orderBook/api"
	//"fmt"
	"strings"
)

type BinanceEvents struct {
	EventType string          `json:"e"`
	TimeStamp int64           `json:"E"`
	Symbol string          `json:"s"`
	FirstId int             `json:"U"`
	FinalId int             `json:"u"`
	Bids [][]string `json:"b"`
	Asks [][]string `json:"a"`
}

type BinanceManager struct {
	CoinManager
	binanceApi     *api.BinanceApi
}

func NewBinanceManager() *BinanceManager {
	var manger = BinanceManager{}
	manger.exchangeBook = newExchangeBook(Binance)
	manger.binanceApi = &api.BinanceApi{}
	return &manger
}

func (b *BinanceManager) StartListen(exchangeConfiguration ExchangeConfiguration, resultChan chan Result) {
	//log.Debugf("StartListen:start binance manager listen")
	ch := make(chan api.Reposponse)
	go b.binanceApi.StartListen(ch)
	go b.startSendingDataBack(exchangeConfiguration, resultChan)

	for {
		select {
		case response := <-ch:

			if *response.Err != nil {
				//log.Errorf("StartListen: binance error:%v", *response.Err)
				exchangeEvents := ExchangeBook{}
				resultChan <- Result{exchangeEvents, response.Err}
			} else if *response.Message != nil {

				Lock.Lock()
				//fmt.Printf("%s \n", *response.Message)
				var binanceOrders BinanceEvents
				json.Unmarshal(*response.Message, &binanceOrders)
				//fmt.Println(b.convert(binanceOrders.Symbol))

				keySymbok := b.convertSymbol(binanceOrders.Symbol)
				//fmt.Println(keySymbok)

				if _, ok := b.exchangeBook.CoinsBooks[keySymbok]; !ok {
					newCoinBook := NewCoinBook(b.convert(binanceOrders.Symbol))
					b.exchangeBook.CoinsBooks[keySymbok] = newCoinBook
				}

				previosCoinBook := b.exchangeBook.CoinsBooks[keySymbok]


				for _, level := range  binanceOrders.Asks {
					price := level[0]
					quantity:= level[1]

					if quantity == "0.00000000" {
						//delete(previosCoinBook.PriceLevels.Asks, price)
						delete(previosCoinBook.Asks, price)
					} else {
						//previosCoinBook.PriceLevels.Asks[price] = quantity
						previosCoinBook.Asks[price] = quantity
					}
				}

				for _, level := range  binanceOrders.Bids {
					price := level[0]
					quantity:= level[1]

					if quantity == "0.00000000" {
						//delete(previosCoinBook.PriceLevels.Bids, price)
						delete(previosCoinBook.Bids, price)
					} else {
						//previosCoinBook.PriceLevels.Bids[price] = quantity
						previosCoinBook.Bids[price] = quantity
					}
				}

				b.exchangeBook.CoinsBooks[keySymbok] = previosCoinBook

				Lock.Unlock()
			} else {
				//log.Errorf("StartListen: Binance mesage is nil")
			}
		}
	}

}


func (b *BinanceManager) convertSymbol(binanceSymbol string) string {

	if len(binanceSymbol) > 0 {
		var symbol = binanceSymbol
		var damagedSymbol = TrimLeftChars(symbol, 1)
		for _, referenceCurrency := range DefaultReferenceCurrencies {
			//fmt.Println(damagedSymbol, referenceCurrency.CurrencyCode())

			if strings.Contains(damagedSymbol, referenceCurrency.CurrencyCode()) {

				//fmt.Println("2",symbol, referenceCurrency.CurrencyCode())
				targetCurrencyString := strings.TrimSuffix(symbol, referenceCurrency.CurrencyCode())

				if targetCurrencyString == "BCC" {
					targetCurrencyString = "BCH"
				}

				return targetCurrencyString+"-"+referenceCurrency.CurrencyCode()
			}
		}

	}
	return ""
}

func (b *BinanceManager) convert(symbol string) CurrencyPair {
	if len(symbol) > 0 {
		var damagedSymbol = TrimLeftChars(symbol, 1)
		for _, referenceCurrency := range DefaultReferenceCurrencies {
			//fmt.Println(damagedSymbol, referenceCurrency.CurrencyCode())

			if strings.Contains(damagedSymbol, referenceCurrency.CurrencyCode()) {

				//fmt.Println("2",symbol, referenceCurrency.CurrencyCode())
				targetCurrencyString := strings.TrimSuffix(symbol, referenceCurrency.CurrencyCode())

				if targetCurrencyString == "BCC" {
					targetCurrencyString = "BCH"
				}

				//fmt.Println(targetCurrencyString)
				var targetCurrency = NewCurrencyWithCode(targetCurrencyString)
				return CurrencyPair{targetCurrency, referenceCurrency}
			}
		}
	}
	return CurrencyPair{NotAplicable, NotAplicable}
}

//type BinanceTicker struct {
//	Symbol             string  `json:"s"`
//	Rate               string  `json:"c"`
//	EventTime          float64 `json:"E"` // field is not needed but it's a workaround because unmarshal is case insensitive and without this filed json can't be parsed
//	StatisticCloseTime float64 `json:"C"` // field is not needed but it's a workaround because unmarshal is case insensitive and without this filed json can't be parsed
//}
//
//func (b *BinanceTicker) getCurriences() CurrencyPair {
//
//	if len(b.Symbol) > 0 {
//		var symbol = b.Symbol
//		var damagedSymbol = TrimLeftChars(symbol, 1)
//		for _, referenceCurrency := range DefaultReferenceCurrencies {
//			//fmt.Println(damagedSymbol, referenceCurrency.CurrencyCode())
//
//			if strings.Contains(damagedSymbol, referenceCurrency.CurrencyCode()) {
//
//				//fmt.Println("2",symbol, referenceCurrency.CurrencyCode())
//				targetCurrencyString := strings.TrimSuffix(symbol, referenceCurrency.CurrencyCode())
//
//				if targetCurrencyString == "BCC" {
//					targetCurrencyString = "BCH"
//				}
//
//				//fmt.Println(targetCurrencyString)
//				var targetCurrency = NewCurrencyWithCode(targetCurrencyString)
//				return CurrencyPair{targetCurrency, referenceCurrency}
//			}
//		}
//
//	}
//	return CurrencyPair{NotAplicable, NotAplicable}
//}