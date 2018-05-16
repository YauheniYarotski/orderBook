package main

import (
	"sync"
	"orderBook/core"
)

var manager = core.NewManager()
var waitGroup = &sync.WaitGroup{}

//var configString = `{
//		"targetCurrencies" : ["BTC", "ETH", "GOLOS", "BTS", "STEEM", "WAVES", "LTC", "BCH", "ETC", "DASH", "EOS"],
//		"referenceCurrencies" : ["USD", "BTC"],
//		"exchanges": ["Binance","Bitfinex","Gdax","HitBtc","Okex","Poloniex"],
//		"refreshInterval" : "3"
//		}`

//const (
//	DbUser     = "postgres"
//	DbPassword = "postgres"
//	DbName     = "test"
//)

func main() {

	var configuration = core.ManagerConfiguration{}
	//, "GOLOS", "BTS", "STEEM", "WAVES", "LTC", "BCH", "ETC", "DASH", "EOS"
	//	[]string{"BTC", "ETH", "GOLOS", "BTS", "STEEM", "WAVES", "LTC", "BCH", "ETC", "DASH", "EOS"}
	//	configuration.TargetCurrencies = []string{"LTC","DASH"}
	configuration.TargetCurrencies = []string{"BTC"}
	configuration.ReferenceCurrencies = []string{"USDT"}
	configuration.Exchanges = []string{"BITFINEX", "BINANCE"}
	configuration.RefreshInterval = 1
//, "Bitfinex", "Gdax", "HitBtc", "Okex", "Poloniex", "Bittrex", "HUOBI", "UPBIT", "KRAKEN", "BITHUMB"
	dbConfig := core.DBConfiguration{}
	dbConfig.User = "postgres"
	dbConfig.Password = "postgres"
	dbConfig.Name = "test"
	configuration.DBConfiguration = dbConfig

	waitGroup.Add(len(configuration.Exchanges) + 8)

	go manager.StartListen(configuration)

	waitGroup.Wait()

}
