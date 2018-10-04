package main

import (
	"orderBook/core"
	"sync"
	"net/http"
	"flag"
	"log"
	"os"
)

var manager = core.NewManager()
var waitGroup = &sync.WaitGroup{}

var addr = flag.String("addr", "0.0.0.0:8080", "http service address")


func main() {

	var configuration = core.ManagerConfiguration{}
	//"BITMEX","BITFINEX", "BINANCE", "BITSTAMP"
	configuration.Exchanges = []string{"BITMEX","BITFINEX", "BINANCE", "BITSTAMP"}
	configuration.RefreshInterval = 1
	dbConfig := core.DBConfiguration{}
	dbConfig.User = "postgres"
	dbConfig.Password = "postgres"
	dbConfig.Name = "test"
	configuration.DBConfiguration = dbConfig


	go manager.Start(configuration)

	http.Handle("/", http.FileServer(http.Dir("./webPages")))

	var httpErr error

	if _, err := os.Stat("./selfsigned.crt"); err == nil {
		log.Println("file ", "selfsigned.crt found switching to https")
		if httpErr = http.ListenAndServeTLS(*addr, "selfsigned.crt", "selfsigned.key", nil); httpErr != nil {
			log.Fatal("The process exited with https error: ", httpErr.Error())
		}
	} else {
		httpErr = http.ListenAndServe(*addr, nil)
		if httpErr != nil {
			log.Fatal("The process exited with http error: ", httpErr.Error())
		}
	}


}
