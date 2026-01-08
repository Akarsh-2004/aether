package engine

import (
	"log"
	"time"

	aetherpb "github.com/yourname/aethersync/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GameEngine is the main engine for the simulation.
type GameEngine struct {
	Registry    *Registry
	Quadtree    *Quadtree
	broadcaster Broadcaster
	ticker      *time.Ticker
	stop        chan bool
	// clientKnowledge tracks what each client last saw.
	// map[clientID]map[entityID]lastSentPosition
	clientKnowledge map[string]map[string]Vec2
}

// NewGameEngine creates a new GameEngine.
func NewGameEngine(registry *Registry, broadcaster Broadcaster, worldBoundary *AABB, quadtreeCapacity int) *GameEngine {
	return &GameEngine{
		Registry:        registry,
		Quadtree:        NewQuadtree(worldBoundary, quadtreeCapacity),
		broadcaster:     broadcaster,
		stop:            make(chan bool),
		clientKnowledge: make(map[string]map[string]Vec2),
	}
}

// Start starts the game engine's tick loop.
func (g *GameEngine) Start(tickRate time.Duration) {
	g.ticker = time.NewTicker(tickRate)
	go func() {
		for {
			select {
			case <-g.ticker.C:
				g.Tick()
			case <-g.stop:
				g.ticker.Stop()
				return
			}
		}
	}()
	log.Println("game engine started")
}

// Stop stops the game engine's tick loop.
func (g *GameEngine) Stop() {
	g.stop <- true
	log.Println("game engine stopped")
}

// Tick performs one tick of the game simulation.
func (g *GameEngine) Tick() {
	// Rebuild the quadtree every tick.
	// For a dynamic world, this is a simple approach.
	// More advanced techniques could update the quadtree instead of rebuilding it.
	g.Quadtree = NewQuadtree(g.Quadtree.boundary, g.Quadtree.capacity)
	entities := g.Registry.All()
	for _, entity := range entities {
		g.Quadtree.Insert(entity)
	}

	// For each entity, find nearby entities and broadcast them.
	for _, entity := range entities {
		// Define an AOI for the entity.
		aoi := &Circle{
			CenterX: entity.Position.X,
			CenterY: entity.Position.Y,
			Radius:  100, // 100 units of radius
		}

		// Query the quadtree for entities in the AOI.
		nearbyEntities := g.Quadtree.QueryCircle(aoi, nil)

		// Get or create client knowledge.
		if _, ok := g.clientKnowledge[entity.ID]; !ok {
			g.clientKnowledge[entity.ID] = make(map[string]Vec2)
		}
		knowledge := g.clientKnowledge[entity.ID]
		newKnowledge := make(map[string]Vec2)

		var entityStates []*aetherpb.EntityState
		var entityDeltas []*aetherpb.MovementDelta

		for _, e := range nearbyEntities {
			lastPos, known := knowledge[e.ID]
			newKnowledge[e.ID] = e.Position

			if !known {
				// New entity in AOI, send full state.
				entityStates = append(entityStates, &aetherpb.EntityState{
					Id:         e.ID,
					Position:   &aetherpb.Vec2{X: float32(e.Position.X), Y: float32(e.Position.Y)},
					Velocity:   &aetherpb.Vec2{X: float32(e.Velocity.X), Y: float32(e.Velocity.Y)},
					Rotation:   float32(e.Rotation),
					LastUpdate: timestamppb.New(e.LastUpdate),
				})
			} else {
				// Known entity, check if it moved.
				if e.Position.X != lastPos.X || e.Position.Y != lastPos.Y {
					entityDeltas = append(entityDeltas, &aetherpb.MovementDelta{
						Id:        e.ID,
						Position:  &aetherpb.Vec2{X: float32(e.Position.X), Y: float32(e.Position.Y)},
						Timestamp: timestamppb.New(e.LastUpdate),
					})
				}
			}
		}

		// Update knowledge to current set of nearby entities.
		g.clientKnowledge[entity.ID] = newKnowledge

		// If nothing to send, skip.
		if len(entityStates) == 0 && len(entityDeltas) == 0 {
			continue
		}

		// Wrap in a WorldSnapshot.
		snapshot := &aetherpb.WorldSnapshot{
			Entities: entityStates,
			Deltas:   entityDeltas,
		}

		// Marshal the snapshot.
		payload, err := proto.Marshal(snapshot)
		if err != nil {
			log.Printf("error marshalling world snapshot: %v", err)
			continue
		}

		// Broadcast the payload to the client associated with the entity.
		if g.broadcaster != nil {
			g.broadcaster.BroadcastTo(entity.ID, payload)
		}
	}
}
