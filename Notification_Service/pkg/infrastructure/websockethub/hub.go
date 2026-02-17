package websockethub

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn   *websocket.Conn
	UserID uint64
}

type Hub struct {
	Clients    map[uint64]*Client
	Register   chan *Client
	Unregister chan uint64
	//broadcast  chan []byte
	Notification chan NotificationMessage
}
type NotificationMessage struct {
	UserID  uint64
	Payload []byte
}

// NewHub initializes the Hub and returns the instance
func NewHub() *Hub {
	return &Hub{
		Clients:      make(map[uint64]*Client),
		Register:     make(chan *Client),
		Unregister:   make(chan uint64),
		Notification: make(chan NotificationMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.Clients[c.UserID] = c
			//fmt.Println("registering", h.clients)

		case id := <-h.Unregister:
			if c, ok := h.Clients[id]; ok {
				c.Conn.Close()
				delete(h.Clients, id)
			}

		// case msg := <-h.broadcast:
		// 	for _, c := range h.clients {
		// 		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		// 		if err != nil {
		// 			c.Conn.Close()
		// 			delete(h.clients, c.UserID)
		// 		}
		// 	}
		case msg := <-h.Notification:
			// FIND the specific user and send ONLY to them
			if client, ok := h.Clients[msg.UserID]; ok {
				err := client.Conn.WriteMessage(websocket.TextMessage, msg.Payload)
				if err != nil {
					client.Conn.Close()
					delete(h.Clients, msg.UserID)
				}
			}
		}
	}
}

// func StartHub() {
// 	//fmt.Println("please do tell me if hub is starting")
// 	go hub.run()
// }
