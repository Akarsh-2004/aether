package engine

// Broadcaster is an interface for broadcasting messages to clients.
type Broadcaster interface {
	BroadcastTo(entityID string, payload []byte)
}
