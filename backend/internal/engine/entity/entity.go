package entity

import (
	"sync"
	"time"

	"github.com/akarsh-2004/aether/proto"
)

type Entity struct {
	ID         uint32
	Type       string
	Position   Vector2
	Velocity   Vector2
	ClientID   string
	LastUpdate time.Time
	
	// Movement validation
	LastSequence uint64
	PendingMoves []*proto.MovementDelta
}

type Vector2 struct {
	X float64
	Y float64
}

type EntityManager struct {
	entities    map[uint32]*Entity
	clientMap   map[string]uint32 // client_id -> entity_id
	nextEntityID uint32
	mu          sync.RWMutex
}

func NewEntityManager() *EntityManager {
	return &EntityManager{
		entities:     make(map[uint32]*Entity),
		clientMap:    make(map[string]uint32),
		nextEntityID: 1,
	}
}

func (em *EntityManager) CreateEntity(entityType string, x, y float64, clientID string) *Entity {
	em.mu.Lock()
	defer em.mu.Unlock()

	entityID := em.nextEntityID
	em.nextEntityID++

	entity := &Entity{
		ID:         entityID,
		Type:       entityType,
		Position:   Vector2{X: x, Y: y},
		Velocity:   Vector2{X: 0, Y: 0},
		ClientID:   clientID,
		LastUpdate: time.Now(),
	}

	em.entities[entityID] = entity
	if clientID != "" {
		em.clientMap[clientID] = entityID
	}

	return entity
}

func (em *EntityManager) GetEntity(id uint32) (*Entity, bool) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	entity, exists := em.entities[id]
	return entity, exists
}

func (em *EntityManager) GetEntityByClient(clientID string) (*Entity, bool) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	entityID, exists := em.clientMap[clientID]
	if !exists {
		return nil, false
	}

	entity, exists := em.entities[entityID]
	return entity, exists
}

func (em *EntityManager) RemoveEntity(id uint32) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	entity, exists := em.entities[id]
	if !exists {
		return false
	}

	delete(em.entities, id)
	if entity.ClientID != "" {
		delete(em.clientMap, entity.ClientID)
	}

	return true
}

func (em *EntityManager) UpdateEntity(id uint32, position, velocity Vector2) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	entity, exists := em.entities[id]
	if !exists {
		return false
	}

	entity.Position = position
	entity.Velocity = velocity
	entity.LastUpdate = time.Now()

	return true
}

func (em *EntityManager) AddPendingMove(id uint32, delta *proto.MovementDelta) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	entity, exists := em.entities[id]
	if !exists {
		return false
	}

	// Only add if sequence is newer than last processed
	if delta.Sequence > entity.LastSequence {
		entity.PendingMoves = append(entity.PendingMoves, delta)
		return true
	}

	return false
}

func (em *EntityManager) GetPendingMoves(id uint32) []*proto.MovementDelta {
	em.mu.Lock()
	defer em.mu.Unlock()

	entity, exists := em.entities[id]
	if !exists {
		return nil
	}

	moves := make([]*proto.MovementDelta, len(entity.PendingMoves))
	copy(moves, entity.PendingMoves)
	return moves
}

func (em *EntityManager) ClearPendingMoves(id uint32, upToSequence uint64) {
	em.mu.Lock()
	defer em.mu.Unlock()

	entity, exists := em.entities[id]
	if !exists {
		return
	}

	// Filter out moves up to the specified sequence
	var remaining []*proto.MovementDelta
	for _, move := range entity.PendingMoves {
		if move.Sequence > upToSequence {
			remaining = append(remaining, move)
		}
	}

	entity.PendingMoves = remaining
	entity.LastSequence = upToSequence
}

func (em *EntityManager) GetAllEntities() []*Entity {
	em.mu.RLock()
	defer em.mu.RUnlock()

	entities := make([]*Entity, 0, len(em.entities))
	for _, entity := range em.entities {
		entities = append(entities, entity)
	}

	return entities
}

func (em *EntityManager) GetEntityCount() int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	return len(em.entities)
}

func (v Vector2) Add(other Vector2) Vector2 {
	return Vector2{X: v.X + other.X, Y: v.Y + other.Y}
}

func (v Vector2) Subtract(other Vector2) Vector2 {
	return Vector2{X: v.X - other.X, Y: v.Y - other.Y}
}

func (v Vector2) Multiply(scalar float64) Vector2 {
	return Vector2{X: v.X * scalar, Y: v.Y * scalar}
}

func (v Vector2) Distance(other Vector2) float64 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	return dx*dx + dy*dy // Return squared distance for performance
}

func (v Vector2) Length() float64 {
	return v.X*v.X + v.Y*v.Y // Return squared length for performance
}
