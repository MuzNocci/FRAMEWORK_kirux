package realtime

import (
	"fmt"
	"math/rand"
	"net/http"

	"golang.org/x/net/websocket"
)

type Client struct {
	id   string
	conn *websocket.Conn
	hub  *Hub
	send chan []byte
}

func newClient(w http.ResponseWriter, r *http.Request, hub *Hub) (*Client, error) {
	var c *Client
	var err error

	handler := websocket.Handler(func(conn *websocket.Conn) {
		c = &Client{
			id:   fmt.Sprintf("%d", rand.Int63()),
			conn: conn,
			hub:  hub,
			send: make(chan []byte, 256),
		}
	})

	server := &websocket.Server{Handler: handler}
	server.ServeHTTP(w, r)

	if c == nil {
		err = fmt.Errorf("websocket upgrade failed")
	}
	return c, err
}

func (c *Client) readPump() {
	defer c.hub.Unregister(c.id)
	var msg []byte
	for {
		if err := websocket.Message.Receive(c.conn, &msg); err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	for msg := range c.send {
		if err := websocket.Message.Send(c.conn, string(msg)); err != nil {
			break
		}
	}
}

func (c *Client) close() {
	close(c.send)
	c.conn.Close()
}
