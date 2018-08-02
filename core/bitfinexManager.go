package core


import (
	"orderBook/api"
	"time"
	//"strconv"
	"encoding/json"
	//"github.com/btcsuite/btcutil"
	"strconv"
	"math"
	"strings"

	//"fmt"
	"fmt"
)

type BitfinexManager struct {
	CoinManager
	bitfinexSymbols map[int]string
	api             *api.BitfinexApi
	restApi *api.RestApi
}

//type BitfinexTicker struct {
//	ChanID     int    `json:"chanId"`
//	Channel    string `json:"channel"`
//	Event      string `json:"event"`
//	Pair       string `json:"pair"`
//	Symbol     string `json:"symbol"`
//	Rate       string
//	TimpeStamp time.Time
//}

type BitfinexBookResponse struct {
	ChanID     int    `json:"chanId"`
	Channel    string `json:"channel"`
	Event      string `json:"event"`
	Pair       string `json:"pair"`
	Len     	string `json:"len"`
	Freq     string `json:"freq"`
	Prec     string `json:"prec"`
	Price      float64
	Count 		float64
	Amount		float64
	TimpeStamp time.Time
}

type BitfinexRestEvents struct {
	events [][]float64
}

func (self *BitfinexManager) StartListen(exchangeConfiguration ExchangeConfiguration, resultChan chan Result) {
	//self.bitfinexTickers = make(map[int]BitfinexTicker)

	self.restApi = api.NewRestApi()
	self.exchangeBook = newExchangeBook(Bitfinex)
	self.bitfinexSymbols = map[int]string{}
	self.api = api.NewBitfinexApi()
	//self.coinBooks = sync.Map{}

	ch := make(chan api.Reposponse)
	restApiResponseChan := make(chan api.RestApiReposponse)

	urlString := "https://api.bitfinex.com/v2/book/tBTCUSD/P0?len=100"
	go self.restApi.PublicRequest(urlString, restApiResponseChan)

	go self.api.StartListen(ch)

	go self.startSendingDataBack(exchangeConfiguration, resultChan)

	for {
		select {
		case response := <-ch:

			if *response.Err != nil {
				//log.Errorf("StartListen *response.Err: %v", response.Err)
				//resultChan <- Result{exchangeConfiguration.Exchange.String(), nil, response.Err}
			} else if *response.Message != nil {
				//fmt.Printf("%s \n", response.Message)
				self.addMessage(*response.Message)
			} else {
				//log.Errorf("StartListen :error parsing Bitfinex ticker")
			}

		case response := <-restApiResponseChan:
			if *response.Err != nil {
				//log.Errorf("StartListen: binance error:%v", *response.Err)
				//exchangeEvents := ExchangeBook{}
				//resultChan <- Result{exchangeEvents, response.Err}
			} else if *response.Message != nil {

				var bitfinexEvents [][]float64
				json.Unmarshal(*response.Message, &bitfinexEvents)

				for _, level := range  bitfinexEvents {
					var pair = "BTC-USDT"
					var price float64
					var count float64
					var amount float64
					price = level[0]
					count = level[1]
					amount = level[2]
					self.addEvent(pair, price, count, amount)
				}

			} else {
				//log.Errorf("StartListen: Binance mesage is nil")
			}


		}
	}

}



func (self *BitfinexManager) addMessage(message []byte) {

	var pair string
	var price float64
	var count float64
	var amount float64



	var bitfinexBook BitfinexBookResponse
	json.Unmarshal(message, &bitfinexBook)



	if bitfinexBook.ChanID > 0 {
		//fmt.Println(self.convertSymbol(bitfinexBook.Pair))
		self.bitfinexSymbols[bitfinexBook.ChanID] = self.convertSymbol(bitfinexBook.Pair)
	} else {
		var unmarshaledBookMessage []interface{}
		json.Unmarshal(message, &unmarshaledBookMessage)
		if len(unmarshaledBookMessage) > 1 {
			var chanId = int(unmarshaledBookMessage[0].(float64))
			//var unmarshaledTicker []interface{}
			//fmt.Println(unmarshaledBookMessage[1])
			if v, ok := unmarshaledBookMessage[1].([]interface{}); ok {

				if len(v) == 3 {

					//fmt.Println(v)

					pair = self.bitfinexSymbols[chanId]
					price = v[0].(float64)
					count = v[1].(float64)
					amount = v[2].(float64)


					//fmt.Println(suself.Price)
					//self.bitfinexTickers[chanId] = sub
				} else if len(v) > 3 {
					for _, vv := range v {
						if events, ok := vv.([]interface{}); ok {
							pair = self.bitfinexSymbols[chanId]
							price = events[0].(float64)
							count = events[1].(float64)
							amount = events[2].(float64)
						}
					}
				}
			}
		}
	}


	if pair != "" {
		self.addEvent(pair, price, count, amount)
	}

}

func (self *BitfinexManager) addEvent(symbol string, price float64, count float64, amount float64)  {

	mu.Lock()

	if _, ok := self.exchangeBook.CoinsBooks[symbol]; !ok {
		newCoinBook := NewCoinBook(self.convertSymbolToPair(symbol))
		self.exchangeBook.CoinsBooks[symbol] = newCoinBook
	}


	previouseCoinBook := self.exchangeBook.CoinsBooks[symbol]

	priceString := strconv.FormatFloat(price, 'f', 4, 64)
	amountString := strconv.FormatFloat(math.Abs(float64(amount)), 'f', 4, 64)

	if count > 0 {

		if amount < 0 {
			previouseCoinBook.Asks[priceString] = amountString
		} else if amount > 0 {
			previouseCoinBook.Bids[priceString] = amountString
		}else {
			fmt.Println("amount can't be:", amount)
		}

	} else if count == 0 {
		if amount == -1 {
			delete(previouseCoinBook.Asks, priceString)
		} else if amount == 1 {
			delete(previouseCoinBook.Bids, priceString)
		} else {
			fmt.Println("amount can't be:", amount)
		}

	} else {
		fmt.Println("count can't be <0:", count)
	}


	self.exchangeBook.CoinsBooks[symbol] = previouseCoinBook
	mu.Unlock()
}
//func (self PoloniexManager) convertArgsToTicker(args []interface{}) (wsticker PoloniexTicker, err error) {
//	wsticker.CurrencyPair = self.channelsByID[strconv.FormatFloat(args[0].(float64), 'f', 0, 64)]
//	wsticker.Last = args[1].(string)
//	return
//}

func (self *BitfinexManager) convertSymbol(binanceSymbol string) string {

	if len(binanceSymbol) > 0 {
		var symbol = binanceSymbol
		var damagedSymbol = TrimLeftChars(symbol, 1)
		for _, referenceCurrency := range DefaultReferenceCurrencies {
			//fmt.Println(damagedSymbol, referenceCurrency.CurrencyCode())

			referenceCurrencyString := referenceCurrency.CurrencyCode()

			if referenceCurrencyString == "USDT" {
				referenceCurrencyString = "USD"
			}

			if strings.Contains(damagedSymbol, referenceCurrencyString) {

				//fmt.Println("2",symbol, referenceCurrency.CurrencyCode())
				targetCurrencyString := strings.TrimSuffix(symbol, referenceCurrencyString)

				return targetCurrencyString+"-"+referenceCurrency.CurrencyCode()
			}
		}

	}
	return ""
}

func (self *BitfinexManager) convertSymbolToPair(symbol string) CurrencyPair {
	if len(symbol) > 0 {
		var damagedSymbol = TrimLeftChars(symbol, 1)
		for _, referenceCurrency := range DefaultReferenceCurrencies {
			//fmt.Println(damagedSymbol, referenceCurrency.CurrencyCode())

			referenceCurrencyCode := referenceCurrency.CurrencyCode()

			if referenceCurrencyCode == "USDT" {
				referenceCurrencyCode = "USD"
			}

			//fmt.Println(damagedSymbol)
			//fmt.Println(referenceCurrencyCode)
			//fmt.Println(strings.Contains(damagedSymbol, referenceCurrencyCode))

			if strings.Contains(damagedSymbol, referenceCurrencyCode) {
				//fmt.Println(damagedSymbol)

				//fmt.Println("2",symbol, referenceCurrency.CurrencyCode())
				//fmt.Println(referenceCurrency.CurrencyCode())

				targetCurrencyString := strings.TrimSuffix(symbol, "-"+referenceCurrency.CurrencyCode())


				if targetCurrencyString == "DSH" {
					targetCurrencyString = "DASH"
				}


				//fmt.Println("targetCurrencyString", targetCurrencyString)
				var targetCurrency = NewCurrencyWithCode(targetCurrencyString)
				return CurrencyPair{ targetCurrency, referenceCurrency}
			}
		}

	}
	return CurrencyPair{NotAplicable, NotAplicable}
}
