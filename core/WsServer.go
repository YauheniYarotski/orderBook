package core


import (
	"net/http"

	"github.com/gorilla/websocket"
	"fmt"
	"html/template"
	"log"

	//"strconv"
	"github.com/bradfitz/slice"
	"time"
	"encoding/json"
	"math"
)

//213.136.80.2




type WsServer struct {
	upgrader websocket.Upgrader
	ServerHandler   func(granulation float64, p *[]ExchangeBook)
	book []WSExchangeBook
	granulation float64

	pattern   string
	//messages  []*Message
	clients   map[int]*Client
	addCh     chan *Client
	delCh     chan *Client
	sendAllCh chan *Message
	doneCh    chan bool
	errCh     chan error
}


func NewWsServer(pattern string) *WsServer {
	log.Println("create WS")
	var ws = WsServer{}
	ws.upgrader = websocket.Upgrader{}
	allowAllOrigin := func(r *http.Request) bool { return true }
	ws.upgrader.CheckOrigin = allowAllOrigin
	ws.granulation = 50

	ws.pattern = pattern
	//ws.messages = []*Message{}
	ws.clients = make(map[int]*Client)
	ws.addCh = make(chan *Client)
	ws.delCh = make(chan *Client)
	ws.sendAllCh = make(chan *Message)
	ws.doneCh = make(chan bool)
	ws.errCh = make(chan error)
	return &ws
}

func (s *WsServer) Add(c *Client) {
	s.addCh <- c
}

func (s *WsServer) Del(c *Client) {
	s.delCh <- c
}

func (s *WsServer) SendAll(msg *Message) {
	s.sendAllCh <- msg
}

func (s *WsServer) Done() {
	s.doneCh <- true
}

func (s *WsServer) Err(err error) {
	s.errCh <- err
}

//func (s *WsServer) sendPastMessages(c *Client) {
//	for _, msg := range s.messages {
//		c.Write(msg)
//	}
//}

func (s *WsServer) sendAll(msg *Message) {
	for _, c := range s.clients {
		if msg.granulation == c.granulation && msg.patern == c.paternt {
			c.Write(msg)
		}
	}
}

func (s *WsServer) start() {

	log.Println("Start WsServer...")
	go s.startSendingAll()
	// websocket handler
	onConnected := func(w http.ResponseWriter, r *http.Request) {

		ws, err := s.upgrader.Upgrade(w, r,  nil)
		if err != nil {
			fmt.Println(err.Error())
			return
		}


		defer func() {
			log.Println("Def close...")
			err := ws.Close()
			if err != nil {
				s.errCh <- err
			}
		}()
		client := NewClient(ws, s, r.RequestURI)
		s.Add(client)
		client.Listen()
	}

	//http.Handle(s.pattern, websocket.Handler(onConnected))
	http.HandleFunc(s.pattern, onConnected)
	http.HandleFunc("/list", onConnected)



	log.Println("Created handler")

	for {
		select {

		// Add new a client
		case c := <-s.addCh:
			log.Println("Added new client")
			s.clients[c.id] = c
			log.Println("Now", len(s.clients), "clients connected.")
			//s.sendPastMessages(c)

			// del a client
		case c := <-s.delCh:
			log.Println("Delete client")
			delete(s.clients, c.id)

			// broadcast message for all clients
		case msg := <-s.sendAllCh:
			//log.Println("Send all:")
			//s.messages = append(s.messages, msg)
			s.sendAll(msg)

		case err := <-s.errCh:
			log.Println("Error:", err.Error())


		case <-s.doneCh:
			log.Println("return got")
			return
		}
	}
}


func (self *WsServer) home(w http.ResponseWriter, r *http.Request) {
	fmt.Println(self.book[0].CoinsBooks[0])
	homeTemplate.Execute(w, self.book[0].CoinsBooks[0])
}

var homeTemplate,_ = template.ParseFiles("./webPages/firstPage.html")

func (self *WsServer) getGranulation() []float64 {
	granulationsMap := map[float64]bool{}
	for _, c := range self.clients {
		granulationsMap[c.granulation] = true
	}
	granulations := make([]float64, 0, len(granulationsMap))
	for k := range granulationsMap {
		granulations = append(granulations, k)
	}
	return granulations
}

func (self *WsServer) startSendingAll() {
	for range time.Tick(1 * time.Second) {
		gr := self.getGranulation()
		for _, granulation := range gr {

			var exchangeBooks []ExchangeBook
			self.ServerHandler(granulation, &exchangeBooks)
			//fmt.Println(exchangeBooks)

			var res []WSExchangeBook

			for _, v := range exchangeBooks {

				newBook := WSExchangeBook{}
				newBook.ExchangeTitle = v.ExchangeTitle

				for k, coinBook := range v.CoinsBooks {
					newCoinBook := NewWsCoinBook(coinBook.Pair)
					newCoinBook.Symbol = k
					totalAsks := 0.0
					for k, v := range coinBook.Asks {
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
					for k, v := range coinBook.Bids {
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

			data, _ := json.Marshal(res)
			message := Message{data, granulation, "/books"}
			self.SendAll(&message)
		}
	}

}





