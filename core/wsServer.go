package core


import (
	"flag"
	"net/http"

	"github.com/gorilla/websocket"
	"fmt"
	"time"
	"encoding/json"
	"github.com/bradfitz/slice"
)

//213.136.80.2

var addr = flag.String("addr", "0.0.0.0:8080", "http service address")



type WsServer struct {
	upgrader websocket.Upgrader
	ServerHandler   func(p *[]ExchangeBook)
}


func NewWsServer() *WsServer {
	fmt.Println("create WS")
	var ws = WsServer{}
	ws.upgrader = websocket.Upgrader{}
	allowAllOrigin := func(r *http.Request) bool { return true }
	ws.upgrader.CheckOrigin = allowAllOrigin
	return &ws
}

func (b *WsServer) books(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("echo")

	c, err := b.upgrader.Upgrade(w, r,  nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer c.Close()

	for range time.Tick(2 * time.Second) {

		var exchangeBooks []ExchangeBook
		b.ServerHandler(&exchangeBooks)
		//fmt.Println(exchangeBooks)

		var res []WSExchangeBook

		for _, v := range exchangeBooks {

			newBook := WSExchangeBook{}
			newBook.ExchangeTitle = v.ExchangeTitle

			for k,coinBook := range v.CoinsBooks {
				newCoinBook := NewCoinBook(coinBook.Pair)
				newCoinBook.Symbol = k

				for k,v := range coinBook.Asks {
					newCoinBook.Asks[k] = v
				}

				newCoinBook.TotalAsks = "555"
				newCoinBook.TotalBids = "777"

				for k,v := range coinBook.Bids {
					newCoinBook.Bids[k] = v
				}



				newBook.CoinsBooks = append(newBook.CoinsBooks, newCoinBook)
			}

			fmt.Println(newBook)
			//res = append(res, newBook)
			res = append(res, newBook)

		}

		slice.Sort(res, func(i, j int) bool {
			return res[i].ExchangeTitle < res[j].ExchangeTitle
		})

		//fmt.Println(res)
		

			//subscribtion := `{"event":"subscribe","channel":"ticker","symbol": ""}`
			msg, _ := json.Marshal(res)
			//fmt.Println(msg)
			err = c.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				//log.Debugf("write:", err)
			}

		}
}


func (b *WsServer) start() {
	//log.Debug("Start WS")
	fmt.Println("start ws")
	flag.Parse()
	http.HandleFunc("/books", b.books)
	//http.HandleFunc("/", home)
	http.Handle("/", http.FileServer(http.Dir("./webPages")))
	http.ListenAndServe(*addr, nil)
	//log.Fatal(http.ListenAndServe(*addr, nil))
}

