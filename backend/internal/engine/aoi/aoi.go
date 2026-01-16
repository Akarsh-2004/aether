package aoi

import (
	"sync"

	"github.com/akarsh-2004/aether/internal/engine/entity"
	"github.com/akarsh-2004/aether/internal/engine/spatial"
)

type AOIManager struct {
	quadtree   *spatial.Quadtree
	aoiRadius  float64
	subscribers map[uint32]map[uint32]struct{} // entity_id -> set of subscriber entity_ids
	mu         sync.RWMutex
}

type AOIEvent struct {
	Type      string    // "enter", "exit", "move"
	EntityID  uint32
	OtherID   uint32
	Position  entity.Vector2
}

func NewAOIManager(quadtree *spatial.Quadtree, aoiRadius float64) *AOIManager {
	return &AOIManager{
		quadtree:    quadtree,
		aoiRadius:   aoiRadius,
		subscribers: make(map[uint32]map[uint32]struct{}),
	}
}

func (am *AOIManager) UpdateEntity(entityID uint32, position entity.Vector2) []AOIEvent {
	am.mu.Lock()
	defer am.mu.Unlock()

	events := make([]AOIEvent, 0)

	// Get current nearby entities
	nearby := am.quadtree.QueryRadius(position, am.aoiRadius)
	
	// Get current subscribers
	currentSubscribers, exists := am.subscribers[entityID]
	if !exists {
		currentSubscribers = make(map[uint32]struct{})
		am.subscribers[entityID] = currentSubscribers
	}

	// Check for new entities (enter events)
	newSubscribers := make(map[uint32]struct{})
	for _, nearbyEntity := range nearby {
		if nearbyEntity.ID == entityID {
			continue // Skip self
		}

		newSubscribers[nearbyEntity.ID] = struct{}{}
		
		if _, wasSubscribed := currentSubscribers[nearbyEntity.ID]; !wasSubscribed {
			events = append(events, AOIEvent{
				Type:     "enter",
				EntityID: entityID,
				OtherID:  nearbyEntity.ID,
				Position: position,
			})
		}
	}

	// Check for entities that are no longer nearby (exit events)
	for otherID := range currentSubscribers {
		if _, stillNearby := newSubscribers[otherID]; !stillNearby {
			events = append(events, AOIEvent{
				Type:     "exit",
				EntityID: entityID,
				OtherID:  otherID,
				Position: position,
			})
		}
	}

	// Update subscribers
	am.subscribers[entityID] = newSubscribers

	// Add move event if there are any subscribers
	if len(newSubscribers) > 0 {
		events = append(events, AOIEvent{
			Type:     "move",
			EntityID: entityID,
			Position: position,
		})
	}

	return events
}

func (am *AOIManager) RemoveEntity(entityID uint32) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Remove from subscribers
	delete(am.subscribers, entityID)

	// Remove from other entities' subscriber lists
	for id, subscribers := range am.subscribers {
		delete(subscribers, entityID)
		am.subscribers[id] = subscribers
	}
}

func (am *AOIManager) GetNearbyEntities(entityID uint32) []uint32 {
	am.mu.RLock()
	defer am.mu.RUnlock()

	subscribers, exists := am.subscribers[entityID]
	if !exists {
		return nil
	}

	nearby := make([]uint32, 0, len(subscribers))
	for id := range subscribers {
		nearby = append(nearby, id)
	}

	return nearby
}

func (am *AOIManager) GetEntitiesInRadius(center entity.Vector2, radius float64) []*entity.Entity {
	return am.quadtree.QueryRadius(center, radius)
}

func (am *AOIManager) GetSubscriberCount() int {
	am.mu.RLock()
	defer am.mu.RUnlock()

	total := 0
	for _, subscribers := range am.subscribers {
		total += len(subscribers)
	}

	return total
}

func (am *AOIManager) GetEntitySubscriberCount(entityID uint32) int {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if subscribers, exists := am.subscribers[entityID]; exists {
		return len(subscribers)
	}

	return 0
}
