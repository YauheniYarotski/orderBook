package main

import (
	"orderBook/core"
	"sync"
)

var manager = core.NewManager()
var waitGroup = &sync.WaitGroup{}



func main() {

	var configuration = core.ManagerConfiguration{}

	configuration.Exchanges = []string{"BITMEX","BITFINEX", "BINANCE", "BITSTAMP"}
	configuration.RefreshInterval = 1
	dbConfig := core.DBConfiguration{}
	dbConfig.User = "postgres"
	dbConfig.Password = "postgres"
	dbConfig.Name = "test"
	configuration.DBConfiguration = dbConfig

	waitGroup.Add(len(configuration.Exchanges) + 2)

	go manager.Start(configuration)

	waitGroup.Wait()

}
