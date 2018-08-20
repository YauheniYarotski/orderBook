
package core

import (
"encoding/json"

	"orderBook/api"
	//"fmt"
	"strings"
	"strconv"
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

type BinanceRestEvents struct {
	LastUpdateID int             `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

type BinanceManager struct {
	CoinManager
	binanceApi     *api.BinanceApi
	restApi *api.RestApi

}

func NewBinanceManager() *BinanceManager {
	var manger = BinanceManager{}
	manger.exchangeBook = newExchangeBook(Binance)
	manger.binanceApi = &api.BinanceApi{}
	manger.restApi = api.NewRestApi()
	return &manger
}

func (self *BinanceManager) StartListen(exchangeConfiguration ExchangeConfiguration, resultChan chan Result) {
	//log.Debugf("StartListen:start binance manager listen")
	ch := make(chan api.Reposponse)
	go self.binanceApi.StartListen(ch)
	go self.startSendingDataBack(exchangeConfiguration, resultChan)

	restApiResponseChan := make(chan api.RestApiReposponse)

	urlString := "https://www.binance.com/api/v1/depth?symbol=BTCUSDT&limit=10000"
	go self.restApi.PublicRequest(urlString, restApiResponseChan)

	for {
		select {
		case response := <-ch:

			if *response.Err != nil {
				//log.Errorf("StartListen: binance error:%v", *response.Err)
				exchangeEvents := ExchangeBook{}
				resultChan <- Result{exchangeEvents, response.Err}
			} else if *response.Message != nil {

				//fmt.Printf("%s \n", *response.Message)
				var binanceEvents BinanceEvents
				json.Unmarshal(*response.Message, &binanceEvents)
				//fmt.Println(b.convert(binanceOrders.Symbol))

				keySymbok := self.convertSymbol(binanceEvents.Symbol)
				//fmt.Println(keySymbok)

				mu.Lock()

				//if map is empty for this pair than, just fill with empty pair
				if _, ok := self.exchangeBook.CoinsBooks[keySymbok]; !ok {
					pair :=  self.convertSymbolToPair(binanceEvents.Symbol)
					newCoinBook := NewCoinBook(pair)
					self.exchangeBook.CoinsBooks[keySymbok] = newCoinBook
				}


				previosCoinBook := self.exchangeBook.CoinsBooks[keySymbok]

				for _, level := range  binanceEvents.Asks {
					price, _ := strconv.ParseFloat(level[0], 64)
					quantity, _ := strconv.ParseFloat(level[1], 64)

					if quantity == 0 {
						//delete(previosCoinBook.PriceLevels.Asks, price)
						delete(previosCoinBook.Asks, price)
					} else {
						//previosCoinBook.PriceLevels.Asks[price] = quantity
						previosCoinBook.Asks[price] = quantity
					}
				}

				for _, level := range  binanceEvents.Bids {
					price, _ := strconv.ParseFloat(level[0], 64)
					quantity, _:= strconv.ParseFloat(level[1], 64)

					if quantity == 0 {
						//delete(previosCoinBook.PriceLevels.Bids, price)
						delete(previosCoinBook.Bids, price)
					} else {
						//previosCoinBook.PriceLevels.Bids[price] = quantity
						previosCoinBook.Bids[price] = quantity
					}
				}

				self.exchangeBook.CoinsBooks[keySymbok] = previosCoinBook
				mu.Unlock()

			} else {
				//log.Errorf("StartListen: Binance mesage is nil")
			}

			//restApi response
		case response := <-restApiResponseChan:

			if *response.Err != nil {
				//log.Errorf("StartListen: binance error:%v", *response.Err)
				exchangeEvents := ExchangeBook{}
				resultChan <- Result{exchangeEvents, response.Err}
			} else if *response.Message != nil {

				//fmt.Printf("%s \n", *response.Message)
				var binanceEvents BinanceRestEvents
				json.Unmarshal(*response.Message, &binanceEvents)
				//fmt.Println(b.convert(binanceOrders.Symbol))

				keySymbok := "BTC-USDT"//self.convertSymbol(binanceEvents.Symbol)
				//fmt.Println(keySymbok)

				mu.Lock()

				//if map is empty for this pair than, just fill with empty pair
				if _, ok := self.exchangeBook.CoinsBooks[keySymbok]; !ok {
					pair :=  self.convertSymbolToPair("BTC-USDT")
					newCoinBook := NewCoinBook(pair)
					self.exchangeBook.CoinsBooks[keySymbok] = newCoinBook
				}


				previosCoinBook := self.exchangeBook.CoinsBooks[keySymbok]

				for _, level := range  binanceEvents.Asks {
					price, _:= strconv.ParseFloat(level[0], 64)
					quantity, _:= strconv.ParseFloat(level[1], 64)

					if quantity == 0 {
						//delete(previosCoinBook.PriceLevels.Asks, price)
						delete(previosCoinBook.Asks, price)
					} else {
						//previosCoinBook.PriceLevels.Asks[price] = quantity
						previosCoinBook.Asks[price] = quantity
					}
				}

				for _, level := range  binanceEvents.Bids {
					price, _:= strconv.ParseFloat(level[0], 64)
					quantity, _:= strconv.ParseFloat(level[1], 64)

					if quantity == 0 {
						//delete(previosCoinBook.PriceLevels.Bids, price)
						delete(previosCoinBook.Bids, price)
					} else {
						//previosCoinBook.PriceLevels.Bids[price] = quantity
						previosCoinBook.Bids[price] = quantity
					}
				}

				self.exchangeBook.CoinsBooks[keySymbok] = previosCoinBook
				mu.Unlock()

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

func (b *BinanceManager) convertSymbolToPair(symbol string) CurrencyPair {
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