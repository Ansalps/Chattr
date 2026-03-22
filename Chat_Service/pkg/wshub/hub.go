package wshub

import (
	"encoding/json"

	"github.com/Ansalps/Chattr_Chat_Service/logger"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/requestmodels"
	"github.com/gorilla/websocket"
)

type Hub struct {
	Clients    map[uint64]*Client
	Register   chan *Client
	Unregister chan uint64
	Broadcast  chan []byte
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

		case msg := <-h.Broadcast:
			for _, c := range h.Clients {
				err := c.Conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					c.Conn.Close()
					delete(h.Clients, c.UserID)
				}
			}
		}
	}
}

func (h *Hub) SendToUser(c *Client, dm requestmodels.MessageRequest, log logger.Logger) error{
	data, err := json.Marshal(dm)
	if err != nil {
		log.Error("failed to marshal individual message",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "sender_id", Value: dm.SenderID},
		)
		SendWSError(c,log,"failed to marshal individual message",h)
		return err
	}
	//SendWSError(c, log, "hi", h)
	client, ok := h.Clients[dm.RecipientID]
	if !ok {
		log.Warn("User offline:",
			logger.Field{Key: "recipient_id", Value: dm.RecipientID})

		 // send ack to sender instead of crashing
		 SendWSError(c, log, "recipient offline", h)

		 return nil // ✅ IMPORTANT
	}
	//data, _ := json.Marshal(dm)
	err = client.Conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Error("failed to send websocket message",
			logger.Field{Key: "user_id", Value: dm.RecipientID},
			logger.Field{Key: "error", Value: err},
		)
		
		// optional: remove dead connection
		h.Unregister <- dm.RecipientID
		client.Conn.Close()
		return err
	}
	return nil
}

func (h *Hub) SendToGroup(c *Client,dm requestmodels.MessageRequest, userIds []uint64, log logger.Logger) error{
	data, err := json.Marshal(dm)
	if err != nil {
		log.Error("failed to marshal group message",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "sender_id", Value: dm.SenderID},
		)
		SendWSError(c,log,"failed to marshal group message",h)
		return err
	}

	for _, v := range userIds {
		if v == dm.SenderID {
			continue
		}
		client, ok := h.Clients[v]
		if !ok {
			log.Warn("User offline:",
				logger.Field{Key: "user_id", Value: v})
			continue
		}
		//data, _ := json.Marshal(dm)
		err := client.Conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Error("failed to send websocket message",
				logger.Field{Key: "user_id", Value: v},
				logger.Field{Key: "error", Value: err},
			)

			// optional: remove dead connection
			h.Unregister <- v
			client.Conn.Close()
			return err
		}
	}
	return nil
}

func SendWSError(c *Client, log logger.Logger, msg string, h *Hub) {
	// resp := map[string]string{
	// 	"type":    "ERROR",
	// 	"message": msg,
	// }

	// data, err := json.Marshal(resp)
	// if err != nil {
	// 	log.Error("failed to marshal error response",
	// 		logger.Field{Key: "error", Value: err},
	// 	)
	// 	if err := c.Conn.WriteMessage(websocket.TextMessage, []byte("failed to marshal error response")); err != nil {
	// 		log.Error("failed to write websocket error",
	// 			logger.Field{Key: "error", Value: err},
	// 		)
	// 	}
	// 	// optional: remove dead connection
	// 	h.Unregister<-c.UserID
	// 	c.Conn.Close()
	// 	return
	// }
	c.WriteMu.Lock()
	 err := c.Conn.WriteMessage(websocket.TextMessage, []byte(msg)); 
	 c.WriteMu.Unlock()
	 if err != nil {
		log.Error("failed to write websocket error",
			logger.Field{Key: "error", Value: err},
		)
		h.Unregister <- c.UserID
		c.Conn.Close()
	}
}
