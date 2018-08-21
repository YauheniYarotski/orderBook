package core

import (
"fmt"
"log"
	"github.com/gorilla/websocket"

	"strconv"
	"io"
)

const channelBufSize = 100

var maxId int = 0

// Chat client.
type Client struct {
	id     int
	ws     *websocket.Conn
	server *WsServer
	ch     chan *Message
	doneCh chan bool
}

// Create new chat client.
func NewClient(ws *websocket.Conn, server *WsServer) *Client {

	if ws == nil {
		panic("ws cannot be nil")
	}

	if server == nil {
		panic("server cannot be nil")
	}

	maxId++
	ch := make(chan *Message, channelBufSize)
	doneCh := make(chan bool)

	return &Client{maxId, ws, server, ch, doneCh}
}

func (c *Client) Conn() *websocket.Conn {
	return c.ws
}

func (c *Client) Write(msg *Message) {
	select {
	case c.ch <- msg:
	default:
		//c.ws.Close()
		c.server.Del(c)
		err := fmt.Errorf("client %d is disconnected.", c.id)
		fmt.Println(err)
		//c.server.Err(err)
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
		case msg := <-c.ch:
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
			//c.ws.Close()
			c.server.Del(c)
			c.doneCh <- true // for listenWrite method
			return

			// read data from websocket connection
		default:
			//var msg Message
			//err := websocket.JSON.Receive(c.ws, &msg)
			//if err == io.EOF {
			//	c.doneCh <- true
			//} else if err != nil {
			//	c.server.Err(err)
			//} else {
			//	c.server.SendAll(&msg)
			//}

			_, message, err := c.ws.ReadMessage()
			if err == io.EOF {
				c.doneCh <- true
			} else if err != nil {
				c.doneCh <- true
				c.server.Err(err)
			} else {

				stringMessage := string(message)
				log.Printf("recv from client %d %@:", c.id, stringMessage)
				granulation, err := strconv.ParseFloat(stringMessage, 64)
				if err == nil {
					c.server.changeGranulation(granulation)
				} else {
					c.server.changeGranulation(50)
				}
			}
		}
	}
}
