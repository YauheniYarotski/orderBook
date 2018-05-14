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
	"fmt"
)

type BitfinexManager struct {
	BasicManager
	bitfinexSymbols map[int]string
	api             *api.BitfinexApi
	cointEvenst CoinEvents
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
	b.bitfinexSymbols = map[int]string{}
	b.api = api.NewBitfinexApi()
	b.cointEvenst = CoinEvents{}

	var apiCurrenciesConfiguration = api.ApiCurrenciesConfiguration{}
	apiCurrenciesConfiguration.TargetCurrencies = exchangeConfiguration.TargetCurrencies
	apiCurrenciesConfiguration.ReferenceCurrencies = exchangeConfiguration.ReferenceCurrencies

	ch := make(chan api.Reposponse)

	go b.api.StartListen(ch)

	go b.startSendingDataBack(exchangeConfiguration, resultChan)

	for {
		select {
		case response := <-ch:

			//fmt.Println(0)
			if *response.Err != nil {
				log.Errorf("StartListen *response.Err: %v", response.Err)
				//resultChan <- Result{exchangeConfiguration.Exchange.String(), nil, response.Err}
			} else if *response.Message != nil {
				//fmt.Printf("%s \n", response.Message)
				//fmt.Println(1)
				b.addMessage(*response.Message)
				//fmt.
			} else {
				log.Errorf("StartListen :error parsing Bitfinex ticker")
			}

		}
	}

}

func (b *BitfinexManager) startSendingDataBack(exchangeConfiguration ExchangeConfiguration, resultChan chan Result) {

	for range time.Tick(1 * time.Second) {
		func() {

			b.Lock()
			tempEvents := CoinEvents{}
			for k, v := range b.cointEvenst {
				k = b.convertSymbol(k)
				fmt.Println(k)
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

func (b *BitfinexManager) addMessage(message []byte) {

	var bitfinexBook BitfinexBookResponse
	json.Unmarshal(message, &bitfinexBook)

	if bitfinexBook.ChanID > 0 {
		//fmt.Println(bitfinexTicker)
		b.Lock()
		b.bitfinexSymbols[bitfinexBook.ChanID] = bitfinexBook.Pair
		b.Unlock()
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


	b.Lock()



	if _, ok := b.cointEvenst[symbol]; !ok {
		b.cointEvenst[symbol] = PriceLevels{make(map[string]string), make(map[string]string)}
	}

	previosPriceLevels := b.cointEvenst[symbol]


	priceString := strconv.FormatFloat(price, 'f', 8, 64)
	amountString := strconv.FormatFloat(math.Abs(float64(amount)), 'f', 8, 64)

	if amount < 0 {
		if amount == 0 {
			delete(previosPriceLevels.Asks, priceString)
		} else {
			previosPriceLevels.Asks[priceString] = amountString
		}


	} else {
		if amount == 0 {
			delete(previosPriceLevels.Bids, priceString)
		} else {
			previosPriceLevels.Bids[priceString] = amountString
		}

	}
	b.Unlock()
	//fmt.Println(b.cointEvenst)
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