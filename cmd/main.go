package main

import (
	"orderBook/core"
	"sync"
	"net/http"
	"flag"
)

var manager = core.NewManager()
var waitGroup = &sync.WaitGroup{}

var addr = flag.String("addr", "0.0.0.0:8080", "http service address")


func main() {

	var configuration = core.ManagerConfiguration{}

	configuration.Exchanges = []string{"BITMEX","BITFINEX", "BINANCE", "BITSTAMP"}
	configuration.RefreshInterval = 1
	dbConfig := core.DBConfiguration{}
	dbConfig.User = "postgres"
	dbConfig.Password = "postgres"
	dbConfig.Name = "test"
	configuration.DBConfiguration = dbConfig


	go manager.Start(configuration)

	http.Handle("/", http.FileServer(http.Dir("./webPages")))

	http.ListenAndServe(*addr, nil)
}
