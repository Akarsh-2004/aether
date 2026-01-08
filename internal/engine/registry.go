package engine

import "sync"

// Registry holds all the entities in the world.
type Registry struct {
	mu       sync.RWMutex
	entities map[string]*Entity
}

// NewRegistry creates a new Registry.
func NewRegistry() *Registry {
	return &Registry{
		entities: make(map[string]*Entity),
	}
}

// Add adds a new entity to the registry.
func (r *Registry) Add(entity *Entity) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entities[entity.ID] = entity
}

// Remove removes an entity from the registry.
func (r *Registry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entities, id)
}

// Get returns an entity by its ID.
func (r *Registry) Get(id string) (*Entity, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entity, ok := r.entities[id]
	return entity, ok
}

// All returns all entities in the registry.
func (r *Registry) All() []*Entity {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]*Entity, 0, len(r.entities))
	for _, entity := range r.entities {
		all = append(all, entity)
	}
	return all
}
