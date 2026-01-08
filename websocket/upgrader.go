package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// upgrader is a websocket upgrader with a permissive CheckOrigin.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// For development, allow all origins.
		// In production, you'd want to check the origin.
		return true
	},
}
