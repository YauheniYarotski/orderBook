package core

import (
"fmt"
"log"
	"github.com/gorilla/websocket"
	"encoding/json"
)

const channelBufSize = 100

var maxId int = 0

// Chat client.
type Client struct {
	id     int
	ws     *websocket.Conn
	server *WsServer
	messageCh     chan *Message
	doneCh chan bool
	granulation float64
	paternt string
}

// Create new chat client.
func NewClient(ws *websocket.Conn, server *WsServer, patern string) *Client {

	if ws == nil {
		panic("ws cannot be nil")
	}

	if server == nil {
		panic("server cannot be nil")
	}

	maxId++
	ch := make(chan *Message)
	doneCh := make(chan bool)

	return &Client{maxId, ws, server, ch, doneCh, 50, patern}
}

func (c *Client) Conn() *websocket.Conn {
	return c.ws
}

func (c *Client) Write(msg *Message) {
	select {
	case c.messageCh <- msg:
	default:
		c.ws.Close()
		c.server.Del(c)
		err := fmt.Errorf("client %d is disconnected.", c.id)
		c.server.Err(err)
	}
}

func (c *Client) Done() {
	c.doneCh <- true
}

// Listen Write and Read request via chanel
func (c *Client) Listen() {
	go c.listenWrite()
	c.listenRead()
}

// Listen write request via chanel
func (c *Client) listenWrite() {
	log.Println("Listening write to client")
	for {
		select {

		// send message to the client
		case msg := <-c.messageCh:
			//log.Println("Send to client:", c.id)
			//websocket.JSON.Send(c.ws, msg)
			err := c.ws.WriteMessage(websocket.TextMessage, msg.Body)
			if err != nil {
				log.Printf("write:", err)
			}

			// receive done request
		case <-c.doneCh:
			//c.ws.Close()
			c.server.Del(c)
			c.doneCh <- true // for listenRead method
			return
		}
	}
}

// Listen read request via chanel
func (c *Client) listenRead() {
	log.Println("Listening read from client")
	for {
		select {

		// receive done request
		case <-c.doneCh:
			c.ws.Close()
			c.server.Del(c)
			c.doneCh <- true // for listenWrite method
			return

			// read data from websocket connection
		default:

			_, message, err := c.ws.ReadMessage()
			granulation := Granulation{}
			json.Unmarshal(message, &granulation)
			if err != nil {

				c.doneCh <- true
				c.server.Err(err)
			} else {

				log.Printf("recv from client %d %d:", c.id, granulation.Granulation)
				if err == nil && granulation.Granulation != 0 {
					c.granulation = granulation.Granulation
				} else {
					c.granulation = 50
				}
			}
		}
	}
}

type Granulation struct {
	Granulation float64 `json:"granulation"`
}
