package core


import (
	"flag"
	"html/template"
	"net/http"

	"github.com/gorilla/websocket"
	"fmt"
	"time"
	"encoding/json"
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
	return &ws
}

func (b *WsServer) books(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("echo")
	c, err := b.upgrader.Upgrade(w, r, nil)
	if err != nil {
		//log.Debugf("upgrade:", err)
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

				for k,v := range coinBook.Bids {
					newCoinBook.Bids[k] = v
				}

				newBook.CoinsBooks = append(newBook.CoinsBooks, newCoinBook)
			}

			res = append(res, newBook)

		}


			//subscribtion := `{"event":"subscribe","channel":"ticker","symbol": ""}`
			msg, _ := json.Marshal(res)
			//fmt.Println(msg)
			err = c.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				//log.Debugf("write:", err)
			}

		}
}

func (b *WsServer)home(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("home")
	homeTemplate.Execute(w, "ws://"+r.Host+"/books")
}

func (b *WsServer) start() {
	//log.Debug("Start WS")
	fmt.Println("start ws")
	flag.Parse()
	http.HandleFunc("/books", b.books)
	http.HandleFunc("/", b.home)
	http.ListenAndServe(*addr, nil)
	//log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))