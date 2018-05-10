
package core

import (
"encoding/json"

	"orderBook/api"
	//"fmt"
	"time"
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

type BinanceManager struct {
	BasicManager
	binanceApi     *api.BinanceApi
	cointEvenst CoinEvents
}

func NewBinanceManager() *BinanceManager {
	var manger = BinanceManager{}
	manger.cointEvenst = CoinEvents{}
	manger.binanceApi = &api.BinanceApi{}
	return &manger
}

func (b *BinanceManager) StartListen(exchangeConfiguration ExchangeConfiguration, resultChan chan Result) {
	log.Debugf("StartListen:start binance manager listen")
	ch := make(chan api.Reposponse)
	go b.binanceApi.StartListen(ch)
	go b.startSendingDataBack(exchangeConfiguration, resultChan)

	for {
		select {
		case response := <-ch:

			if *response.Err != nil {
				log.Errorf("StartListen: binance error:%v", *response.Err)
				exchangeEvents := ExchangeEvents{}
				resultChan <- Result{exchangeEvents, response.Err}
			} else if *response.Message != nil {
				//fmt.Printf("%s \n", *response.Message)
				var binanceOrders BinanceEvents
				json.Unmarshal(*response.Message, &binanceOrders)
				//fmt.Println(binanceOrders.Bids)

				if _, ok := b.cointEvenst[binanceOrders.Symbol]; !ok {
					b.cointEvenst[binanceOrders.Symbol] = PriceLevels{make(map[string]string), make(map[string]string)}
				}

				previosPriceLevels := b.cointEvenst[binanceOrders.Symbol]

				for _, level := range  binanceOrders.Asks {
					price := level[0]
					quantity:= level[1]

					if quantity == "0.00000000" {
						delete(previosPriceLevels.Asks, price)
					} else {
						previosPriceLevels.Asks[price] = quantity
					}


				}

				for _, level := range  binanceOrders.Bids {
					price := level[0]
					quantity:= level[1]

					if quantity == "0.00000000" {
						delete(previosPriceLevels.Bids, price)
					} else {
						previosPriceLevels.Bids[price] = quantity
					}

				}


				//fmt.Println(previosPriceLevels, "\n")
				//fmt.Println(len(b.BidsAsks[binanceOrders.Symbol].Bids), "\n")

				b.cointEvenst[binanceOrders.Symbol] = previosPriceLevels

			} else {
				log.Errorf("StartListen: Binance mesage is nil")
			}
		}
	}

}


func (b *BinanceManager) startSendingDataBack(exchangeConfiguration ExchangeConfiguration, resultChan chan Result) {

	for range time.Tick(1 * time.Second) {
		func() {

			b.Lock()
			tempEvents := CoinEvents{}
			for k, v := range b.cointEvenst {
				k = b.convertSymbol(k)
				tempEvents[k] = v
			}
			b.Unlock()



			//fmt.Println(tickerCollection)
			if len(tempEvents) > 0 {
				exchangeEvents := ExchangeEvents{}
				exchangeEvents[exchangeConfiguration.Exchange.String()] = tempEvents
				resultChan <- Result{exchangeEvents, nil}
			}
		}()
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