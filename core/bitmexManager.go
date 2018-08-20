package core


import (
	"orderBook/api"
	//"strconv"
	"encoding/json"
	//"github.com/btcsuite/btcutil"
	"strings"

	//"fmt"
	"fmt"
)

type BitmexManager struct {
	CoinManager
	bitfinexSymbols map[int]string
	bitMexIds map[int64]BitmexData
	api             *api.BitmexApi
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

type BitmexBookResponse struct {
	Table  string `json:"table"`
	Action string `json:"action"`
	Data   []BitmexData `json:"data"`
}

type BitmexData struct {
	Symbol string `json:"symbol"`
	ID     int64  `json:"id"`
	Side   string `json:"side"`
	Size   float64     `json:"size"`
	Price  float64    `json:"price"`
}

type BitmexRestEvents struct {
	events [][]float64
}

func (self *BitmexManager) StartListen(exchangeConfiguration ExchangeConfiguration, resultChan chan Result) {
	//self.bitfinexTickers = make(map[int]BitfinexTicker)

	self.restApi = api.NewRestApi()
	self.exchangeBook = newExchangeBook(Bitfinex)
	self.bitfinexSymbols = map[int]string{}
	self.bitMexIds = map[int64]BitmexData{}
	self.api = api.NewBitmexApi()
	//self.coinBooks = sync.Map{}

	ch := make(chan api.Reposponse)
	restApiResponseChan := make(chan api.RestApiReposponse)

	urlString := "https://api.bitfinex.com/v2/book/tBTCUSD/P2?len=100"
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
				fmt.Println("StartListen :error parsing Bitfinex ticker")
			}

		case response := <-restApiResponseChan:
			if *response.Err != nil {
				//log.Errorf("StartListen: binance error:%v", *response.Err)
				//exchangeEvents := ExchangeBook{}
				//resultChan <- Result{exchangeEvents, response.Err}
			} else if *response.Message != nil {

				var bitfinexEvents [][]float64
				json.Unmarshal(*response.Message, &bitfinexEvents)

				//for _, level := range  bitfinexEvents {
				//	var pair = "BTC-USDT"
				//	var price float64
				//	var count float64
				//	var amount float64
				//	price = level[0]
				//	count = level[1]
				//	amount = level[2]
				//	//self.addEvent(pair, price, count, amount)
				//}

			} else {
				//log.Errorf("StartListen: Binance mesage is nil")
			}


		}
	}

}



func (self *BitmexManager) addMessage(message []byte) {



	var bitmexBook BitmexBookResponse
	json.Unmarshal(message, &bitmexBook)
	self.addEvent(bitmexBook)

	//if pair != "" {
	//	self.addEvent(pair, price, count, amount)
	//}

}

func (self *BitmexManager) addEvent(reponse BitmexBookResponse)  {

	mu.Lock()

	for _,level := range reponse.Data {


		if _, ok := self.exchangeBook.CoinsBooks[level.Symbol]; !ok {
			newCoinBook := NewCoinBook(self.convertSymbolToPair(level.Symbol))
			self.exchangeBook.CoinsBooks[level.Symbol] = newCoinBook
		}

		previouseCoinBook := self.exchangeBook.CoinsBooks[level.Symbol]


		switch action := reponse.Action; action {
		case  "partial", "insert":
			self.bitMexIds[level.ID] = level

			if level.Side == "Buy" {
				previouseCoinBook.Bids[level.Price] = level.Size
			} else if level.Side == "Sell" {
				//if level.Price < 6430 {
				//	fmt.Println("insert:",level.Price)
				//}
				previouseCoinBook.Asks[level.Price] = level.Size
			}

		case "delete":

			leveltToDelete := self.bitMexIds[level.ID]


			if leveltToDelete.Side == "Buy" {
				delete(previouseCoinBook.Bids, leveltToDelete.Price)
			} else if leveltToDelete.Side == "Sell" {
				delete(previouseCoinBook.Asks, leveltToDelete.Price)
				//if level.Price < 6430 {
				//	fmt.Println("delete:", leveltToDelete.Price)
				//}
			}

			delete(self.bitMexIds, level.ID)



		case "update":

			leveltToUpdate := self.bitMexIds[level.ID]


			leveltToUpdate.Side = level.Side
			leveltToUpdate.Size = level.Size

			if leveltToUpdate.Side == "Buy" {
				previouseCoinBook.Bids[leveltToUpdate.Price] = leveltToUpdate.Size
			} else if leveltToUpdate.Side == "Sell" {
				previouseCoinBook.Asks[leveltToUpdate.Price] = leveltToUpdate.Size
				//if level.Price < 6430 {
				//	fmt.Println("update:",leveltToUpdate.Price)
				//}
			}


			self.bitMexIds[level.ID] = leveltToUpdate

		default:
			fmt.Printf("unknown action: ", action)
		}




		self.exchangeBook.CoinsBooks[level.Symbol] = previouseCoinBook
	}
	mu.Unlock()
}
//func (self PoloniexManager) convertArgsToTicker(args []interface{}) (wsticker PoloniexTicker, err error) {
//	wsticker.CurrencyPair = self.channelsByID[strconv.FormatFloat(args[0].(float64), 'f', 0, 64)]
//	wsticker.Last = args[1].(string)
//	return
//}

func (self *BitmexManager) convertSymbol(binanceSymbol string) string {

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

func (self *BitmexManager) convertSymbolToPair(symbol string) CurrencyPair {
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
