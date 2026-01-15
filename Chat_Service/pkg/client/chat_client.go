package client

import "github.com/gorilla/websocket"

type Client struct {
	Conn   *websocket.Conn
	UserID uint64
}
