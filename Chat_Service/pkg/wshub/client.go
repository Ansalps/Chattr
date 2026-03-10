package wshub

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn      *websocket.Conn
	UserID    uint64
	RequestID string
	WriteMu   sync.Mutex
}
