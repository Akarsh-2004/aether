package engine

import "time"

// Vec2 represents a 2D vector.
type Vec2 struct {
	X float64
	Y float64
}

// Entity represents an object in the world.
// This is a temporary definition until we can generate it from protobuf.
type Entity struct {
	ID         string
	Position   Vec2
	Velocity   Vec2
	Rotation   float64
	LastUpdate time.Time
}

// ClientInput is a message from the client to update its entity's state.
type ClientInput struct {
	VelocityX float64 `json:"vx"`
	VelocityY float64 `json:"vy"`
}
