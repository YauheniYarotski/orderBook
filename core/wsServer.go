package core


import (
	"flag"
	"net/http"

	"github.com/gorilla/websocket"
	"fmt"
	"time"
	"encoding/json"
	"github.com/bradfitz/slice"
	"html/template"
	"math"

)

//213.136.80.2

var addr = flag.String("addr", "0.0.0.0:8080", "http service address")



type WsServer struct {
	upgrader websocket.Upgrader
	ServerHandler   func(p *[]ExchangeBook)
	book []WSExchangeBook
}


func NewWsServer() *WsServer {
	fmt.Println("create WS")
	var ws = WsServer{}
	ws.upgrader = websocket.Upgrader{}
	allowAllOrigin := func(r *http.Request) bool { return true }
	ws.upgrader.CheckOrigin = allowAllOrigin
	return &ws
}

func (self *WsServer) books(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("echo")

	c, err := self.upgrader.Upgrade(w, r,  nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer c.Close()

	for range time.Tick(1 * time.Second) {

		var exchangeBooks []ExchangeBook
		self.ServerHandler(&exchangeBooks)
		//fmt.Println(exchangeBooks)

		var res []WSExchangeBook

		for _, v := range exchangeBooks {

			newBook := WSExchangeBook{}
			newBook.ExchangeTitle = v.ExchangeTitle

			for k,coinBook := range v.CoinsBooks {
				newCoinBook := NewWsCoinBook(coinBook.Pair)
				newCoinBook.Symbol = k
				totalAsks := 0.0
				for k,v := range coinBook.Asks {
					if v >= 1 {
						newCoinBook.Asks = append(newCoinBook.Asks, []float64{k, math.Round(v)})
					//newCoinBook.Asks = append(newCoinBook.Asks, []float64{k,v})
						totalAsks = totalAsks + v
					}
				}

				slice.Sort(newCoinBook.Asks, func(i, j int) bool {
					return newCoinBook.Asks[i][0] < newCoinBook.Asks[j][0]
				})

				newCoinBook.TotalAsks = math.Trunc(totalAsks)


				totalBids := 0.0
				for k,v := range coinBook.Bids {
					if v >= 1 {
						newCoinBook.Bids = append(newCoinBook.Bids, []float64{k, math.Round(v)})
						totalBids = totalBids + v
					}
				}
				newCoinBook.TotalBids = math.Trunc(totalBids)

				slice.Sort(newCoinBook.Bids, func(i, j int) bool {
					return newCoinBook.Bids[i][0] > newCoinBook.Bids[j][0]
				})


				newBook.CoinsBooks = append(newBook.CoinsBooks, newCoinBook)
			}

			//fmt.Println(newBook)
			//res = append(res, newBook)
			res = append(res, newBook)

		}

		slice.Sort(res, func(i, j int) bool {
			return res[i].ExchangeTitle < res[j].ExchangeTitle
		})

		//fmt.Println(res)
		//}

		//self.book = res

			//subscribtion := `{"event":"subscribe","channel":"ticker","symbol": ""}`
			msg, _ := json.Marshal(res)
			//fmt.Println(msg)
			err = c.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				//log.Debugf("write:", err)
			}

		}
}


func (self *WsServer) start() {
	//log.Debug("Start WS")
	fmt.Println("start ws")
	flag.Parse()
	http.HandleFunc("/books", self.books)
	//http.HandleFunc("/", self.home)
	http.Handle("/", http.FileServer(http.Dir("./webPages")))
	http.ListenAndServe(*addr, nil)
	//log.Fatal(http.ListenAndServe(*addr, nil))
}

func (self *WsServer) home(w http.ResponseWriter, r *http.Request) {
	fmt.Println(self.book[0].CoinsBooks[0])
	homeTemplate.Execute(w, self.book[0].CoinsBooks[0])
}

var homeTemplate,_ = template.ParseFiles("./webPages/firstPage.html")



