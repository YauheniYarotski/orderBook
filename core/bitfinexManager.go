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
	"sync"
	//"fmt"
)

type BitfinexManager struct {
	CoinManager
	bitfinexSymbols map[int]string
	api             *api.BitfinexApi
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

func (b *BitfinexManager) StartListen(exchangeConfiguration ExchangeConfiguration, resultChan chan Result) {
	//b.bitfinexTickers = make(map[int]BitfinexTicker)

	b.exchangeBook = ExchangeBook{}
	b.exchangeBook.Exchange = Bitfinex
	b.exchangeBook.Coins = sync.Map{}
	b.bitfinexSymbols = map[int]string{}
	b.api = api.NewBitfinexApi()
	//b.coinBooks = sync.Map{}

	var apiCurrenciesConfiguration = api.ApiCurrenciesConfiguration{}
	apiCurrenciesConfiguration.TargetCurrencies = exchangeConfiguration.TargetCurrencies
	apiCurrenciesConfiguration.ReferenceCurrencies = exchangeConfiguration.ReferenceCurrencies

	ch := make(chan api.Reposponse)

	go b.api.StartListen(ch)

	go b.startSendingDataBack(exchangeConfiguration, resultChan)

	for {
		select {
		case response := <-ch:

			if *response.Err != nil {
				log.Errorf("StartListen *response.Err: %v", response.Err)
				//resultChan <- Result{exchangeConfiguration.Exchange.String(), nil, response.Err}
			} else if *response.Message != nil {
				//fmt.Printf("%s \n", response.Message)
				b.addMessage(*response.Message)
			} else {
				log.Errorf("StartListen :error parsing Bitfinex ticker")
			}

		}
	}

}

func (b *BitfinexManager) startSendingDataBack(exchangeConfiguration ExchangeConfiguration, resultChan chan Result) {

	for range time.Tick(1 * time.Second) {
		func() {
			//b.exchangeBook.Coins.Range(func(key, value interface{}) bool {
			//	fmt.Println(value.(CoinBook))
			//	return true
			//})
			resultChan <- Result{b.exchangeBook, nil}
		}()
	}
}

func (b *BitfinexManager) addMessage(message []byte) {

	var bitfinexBook BitfinexBookResponse
	json.Unmarshal(message, &bitfinexBook)

	if bitfinexBook.ChanID > 0 {
		//fmt.Println(b.convertSymbol(bitfinexBook.Pair))
		b.bitfinexSymbols[bitfinexBook.ChanID] = b.convertSymbol(bitfinexBook.Pair)
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

					pair := b.bitfinexSymbols[chanId]
					price := v[0].(float64)
					count := v[1].(float64)
					amount := v[2].(float64)

					b.addEvent(pair, price, count, amount)

					//fmt.Println(sub.Price)
					//b.bitfinexTickers[chanId] = sub
				} else if len(v) > 3 {
					for _, vv := range v {
						if events, ok := vv.([]interface{}); ok {
							pair := b.bitfinexSymbols[chanId]
							price := events[0].(float64)
							count := events[1].(float64)
							amount := events[2].(float64)
							b.addEvent(pair, price, count, amount)
						}
					}
				}
			}
		}
	}
}

func (b *BitfinexManager) addEvent(symbol string, price float64, count float64, amount float64)  {

	newCoinBook := CoinBook{}
	newCoinBook.Pair = b.convert(symbol)
	//fmt.Println(b.convert(symbol))
	newCoinBook.PriceLevels = PriceLevels{sync.Map{}, sync.Map{}}

	previouseCoinBookI, _ := b.exchangeBook.Coins.LoadOrStore(symbol, newCoinBook)
	previouseCoinBook := previouseCoinBookI.(CoinBook)

	priceString := strconv.FormatFloat(price, 'f', 8, 64)
	amountString := strconv.FormatFloat(math.Abs(float64(amount)), 'f', 8, 64)

	if amount < 0 {
		if amount == 0 {
			//delete(coinBook.PriceLevels.Asks, priceString)
			previouseCoinBook.PriceLevels.Asks.Delete(priceString)

		} else {
			//coinBook.PriceLevels.Asks[priceString] = amountString
			previouseCoinBook.PriceLevels.Asks.Store(priceString, amountString)
		}


	} else {
		if amount == 0 {
			//delete(coinBook.PriceLevels.Bids, priceString)
			previouseCoinBook.PriceLevels.Bids.Delete(priceString)
		} else {
			//coinBook.PriceLevels.Bids[priceString] = amountString
			previouseCoinBook.PriceLevels.Bids.Store(priceString, amountString)
		}

	}

	b.exchangeBook.Coins.Store(symbol, previouseCoinBook)
	//b.coinBooks[symbol] = coinBook
	//b.Unlock()
	//fmt.Println(coinBook)
	//fmt.Println(b.coinBooks)
}
//func (b PoloniexManager) convertArgsToTicker(args []interface{}) (wsticker PoloniexTicker, err error) {
//	wsticker.CurrencyPair = b.channelsByID[strconv.FormatFloat(args[0].(float64), 'f', 0, 64)]
//	wsticker.Last = args[1].(string)
//	return
//}

func (b *BitfinexManager) convertSymbol(binanceSymbol string) string {

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

func (b *BitfinexManager) convert(symbol string) CurrencyPair {
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
