package websocket

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/yourname/aethersync/internal/engine"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// entityRegistry is the in-memory store for all entities.
	entityRegistry *engine.Registry

	// clientsByEntityID maps entity IDs to clients.
	clientsByEntityID map[string]*Client

	// nextEntityID is used to assign unique IDs to new entities.
	nextEntityID uint64
}

// NewHub creates a new Hub.
func NewHub(registry *engine.Registry) *Hub {
	return &Hub{
		broadcast:         make(chan []byte),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		clients:           make(map[*Client]bool),
		entityRegistry:    registry,
		clientsByEntityID: make(map[string]*Client),
	}
}

// BroadcastTo sends a message to a specific client by its entity ID.
func (h *Hub) BroadcastTo(entityID string, payload []byte) {
	if client, ok := h.clientsByEntityID[entityID]; ok {
		select {
		case client.send <- payload:
		default:
			// Backpressure: If the client's send buffer is full,
			// we drop this update for this client to prevent head-of-line blocking.
			// We don't disconnect yet, but we log it if it happens too much.
		}
	}
}

// Run starts the hub's event loop.
func (h *Hub) Run() {
	log.Println("hub is running")
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			entityID := atomic.AddUint64(&h.nextEntityID, 1)
			client.entityID = fmt.Sprintf("entity-%d", entityID)
			h.clientsByEntityID[client.entityID] = client
			entity := &engine.Entity{
				ID:         client.entityID,
				Position:   engine.Vec2{X: 0, Y: 0}, // Default position
				LastUpdate: time.Now(),
			}
			h.entityRegistry.Add(entity)
			log.Printf("client registered with entity ID %s", client.entityID)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.clientsByEntityID, client.entityID)
				close(client.send)
				h.entityRegistry.Remove(client.entityID)
				log.Printf("client with entity ID %s unregistered", client.entityID)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					delete(h.clientsByEntityID, client.entityID)
					h.entityRegistry.Remove(client.entityID)
				}
			}
		}
	}
}
