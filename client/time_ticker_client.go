package client

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/caelifer/webtime/service"
	"github.com/gorilla/websocket"
)

// Public package variable to contol Websocket buffers
var WSUpgrader = &websocket.Upgrader{ReadBufferSize: 256, WriteBufferSize: 256}

type WSClient struct {
	service service.Service
}

func (cl *WSClient) log(msg string, args ...interface{}) {
	m := "[CLIENT] " + msg
	log.Printf(m, args...)
}

func NewWSClient(srv service.Service) *WSClient {
	return &WSClient{service: srv}
}

func (cl *WSClient) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get remote address
	peer := r.RemoteAddr

	// Create websocket
	ws, err := WSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		cl.log("(%s) - Failed to create WebSocket, exiting\n", peer)
		return // Need to better handle errors
	}

	// Connection object
	c := &connection{
		send: make(chan []byte, 256),
		ws:   ws,
	}

	// Init service

	// Make quit channel
	quit := make(chan bool)

	// Get service channel
	servchan := cl.service.Service(quit)

	cl.log("(%s) - initialized serivce\n", peer)

	// Loop in the separate goroutine
	go func() {
		for {
			select {
			case update := <-servchan:
				ts := update.(string) // Cast to the correct type
				cl.log("(%s) - sending update [%v]\n", peer, ts)

				msg, err := json.Marshal(&timeUpdate{ts})
				if err != nil {
					cl.log("Bad encoding for %q [%v]\n", ts, err)
					return
				}
				c.send <- []byte(msg)

			case <-quit:
				cl.log("(%s)- got quit signal\n", peer)
				close(quit)
				return
			}
		}
	}()

	// Enable by-directional communication via WS
	go c.write()

	// Reader will return when WS is closed on the remote client-side
	c.read()

	// Clean-up
	quit <- true
}

type connection struct {
	send chan []byte
	ws   *websocket.Conn
}

func (c *connection) read() {
	for {
		_, _, err := c.ws.ReadMessage()
		if err != nil {
			// Close on error
			break
		}
	}
	c.ws.Close()
}

func (c *connection) write() {
	for msg := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println(err)
			c.ws.Close()
			break
		}
	}
}

// Helper type to allow for fast JSON serialization
type timeUpdate struct {
	Time string `json:"time"`
}
