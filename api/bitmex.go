package api


import (
"net/url"

"fmt"

"github.com/gorilla/websocket"
)
const bitmexHost = "www.bitmex.com"
const bitmexPath = "/realtime?subscribe=instrument"

//type biftfinexSubscription struct {
//	Command string `json:"event"`
//	Channel string `json:"channel"`
//	Pair  string `json:"pair"`
//	Prec  string `json:"prec"`
//}

type BitmexApi struct {
	connection           *websocket.Conn
	symbolesForSubscirbe []string
}



func NewBitmexApi() *BitmexApi {
	var api = BitmexApi{}
	//api.symbolesForSubscirbe = []string{"tBTCUSD", "tETHUSD","tBTSUSD", "tSTEEMUSD", "tWAVESUSD", "tLTCUSD", "tBCHUSD", "tETCUSD", "tDASHUSD", "tEOSUSD",  "tETHBTC","tBTSBTC", "tSTEEMBTC", "tWAVESBTC", "tLTCBTC", "tBCHBTC", "tETCBTC", "tDASHBTC", "tEOSBTC"}
	return &api
}

func (b *BitmexApi) connectWs() *websocket.Conn {
	url := url.URL{Scheme: "wss", Host: bitmexHost, Path: bitmexPath}
	//log.Printf("connecting to %s", url.String())

	connection, _, err := websocket.DefaultDialer.Dial(url.String(), nil)

	if err != nil || connection == nil {
		//log.Errorf("connectWs:Bitfinex ws connection error: ", err)
		return nil
	} else {
		//log.Debugf("connectWs:Bitfinex ws connected")
		//b.symbolesForSubscirbe = b.composeSymbolsForSubscirbe(apiCurrenciesConfiguration)
		//for _, symbol := range b.symbolesForSubscirbe {
		subscribtion := `{"op": "subscribe", "args": ["orderBookL2:XBTUSD"]}`
		//fmt.Println(subscribtion)
		connection.WriteMessage(websocket.TextMessage, []byte(subscribtion))
		//}
		return connection
	}
}

func (b *BitmexApi) StartListen(ch chan Reposponse) {
	fmt.Println("StartListen:Start listen Bitmex")
	for {
		if b.connection == nil {
			b.connection = b.connectWs()
		} else if b.connection != nil {

			func() {
				_, message, err := b.connection.ReadMessage()
				if err != nil {
					//log.Errorf("StartListen:Bitfinex read message error: %v", err.Error())
					fmt.Println("close")
					b.connection.Close()
					b.connection = nil
				} else {
					//fmt.Printf("%s \n", message)
					ch <- Reposponse{Message: &message, Err: &err}
				}
			}()
		}
	}

}



func (b *BitmexApi) StopListen() {
	if b.connection != nil {
		b.connection.Close()
		b.connection = nil
	}
	//log.Debugf("Bitfinex ws closed")
}
