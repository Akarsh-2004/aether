package engine

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/akarsh-2004/aether/internal/config"
	"github.com/akarsh-2004/aether/internal/engine/aoi"
	"github.com/akarsh-2004/aether/internal/engine/entity"
	"github.com/akarsh-2004/aether/internal/engine/spatial"
	"github.com/akarsh-2004/aether/internal/engine/tick"
	"github.com/akarsh-2004/aether/proto"
	"go.uber.org/zap"
)

type SpatialEngine struct {
	config         config.EngineConfig
	logger         *zap.Logger
	tickManager    *tick.TickManager
	entityManager  *entity.EntityManager
	quadtree       *spatial.Quadtree
	aoiManager     *aoi.AOIManager
	movementBuffer map[uint32][]*proto.MovementDelta
	mu             sync.RWMutex
	broadcastChan  chan BroadcastMessage
	shutdown       chan struct{}
	wg             sync.WaitGroup
}

type BroadcastMessage struct {
	ClientID string
	Data     []byte
}

func NewSpatialEngine(cfg config.EngineConfig, logger *zap.Logger) *SpatialEngine {
	se := &SpatialEngine{
		config:         cfg,
		logger:         logger,
		entityManager:  entity.NewEntityManager(),
		quadtree:       spatial.NewQuadtreeFromConfig(cfg),
		aoiManager:     aoi.NewAOIManager(se.quadtree, cfg.AOIRadius),
		movementBuffer: make(map[uint32][]*proto.MovementDelta),
		broadcastChan:  make(chan BroadcastMessage, 1000),
		shutdown:       make(chan struct{}),
	}

	se.tickManager = tick.NewTickManager(cfg, logger)
	se.tickManager.AddHandler(se)

	return se
}

func (se *SpatialEngine) Start(ctx context.Context) error {
	se.wg.Add(1)
	go se.broadcastWorker()

	return se.tickManager.Start(ctx)
}

func (se *SpatialEngine) Shutdown(ctx context.Context) error {
	se.tickManager.Shutdown()
	close(se.shutdown)
	
	done := make(chan struct{})
	go func() {
		se.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (se *SpatialEngine) OnTick(tickNumber uint64) {
	se.processTick(tickNumber)
}

func (se *SpatialEngine) OnShutdown() {
	se.logger.Info("Spatial engine shutting down")
}

func (se *SpatialEngine) processTick(tickNumber uint64) {
	start := time.Now()
	
	se.mu.Lock()
	defer se.mu.Unlock()

	// Process all pending movement deltas
	se.processMovementDeltas()

	// Update entity positions based on velocity
	se.updateEntityPositions()

	// Update spatial index
	se.updateSpatialIndex()

	// Process AOI events and generate broadcasts
	se.processAOIEvents()

	duration := time.Since(start)
	if duration > time.Duration(se.config.TickRateMs/2)*time.Millisecond {
		se.logger.Warn("Engine tick processing slow",
			zap.Uint64("tick", tickNumber),
			zap.Duration("duration", duration),
		)
	}
}

func (se *SpatialEngine) SpawnEntity(entityType string, x, y float64, clientID string) uint32 {
	// Validate spawn position
	if !se.isPositionValid(x, y) {
		se.logger.Warn("Invalid spawn position",
			zap.String("client_id", clientID),
			zap.Float64("x", x),
			zap.Float64("y", y),
		)
		return 0
	}

	ent := se.entityManager.CreateEntity(entityType, x, y, clientID)
	if ent == nil {
		return 0
	}

	// Insert into spatial index
	if !se.quadtree.Insert(ent) {
		se.entityManager.RemoveEntity(ent.ID)
		se.logger.Error("Failed to insert entity into spatial index", zap.Uint32("entity_id", ent.ID))
		return 0
	}

	se.logger.Info("Entity spawned",
		zap.Uint32("entity_id", ent.ID),
		zap.String("entity_type", entityType),
		zap.String("client_id", clientID),
		zap.Float64("x", x),
		zap.Float64("y", y),
	)

	return ent.ID
}

func (se *SpatialEngine) RemoveEntity(entityID uint32) bool {
	se.mu.Lock()
	defer se.mu.Unlock()

	ent, exists := se.entityManager.GetEntity(entityID)
	if !exists {
		return false
	}

	// Remove from spatial index
	se.quadtree.Remove(ent)

	// Remove from entity manager
	se.entityManager.RemoveEntity(entityID)

	// Remove from AOI manager
	se.aoiManager.RemoveEntity(entityID)

	// Clear movement buffer
	delete(se.movementBuffer, entityID)

	se.logger.Info("Entity removed", zap.Uint32("entity_id", entityID))
	return true
}

func (se *SpatialEngine) ProcessMovementIntent(entityID uint32, delta *proto.MovementDelta) {
	se.mu.Lock()
	defer se.mu.Unlock()

	// Validate movement
	if !se.validateMovement(entityID, delta) {
		se.logger.Warn("Invalid movement detected",
			zap.Uint32("entity_id", entityID),
			zap.Uint64("sequence", delta.Sequence),
			zap.Float64("delta_x", delta.DeltaX),
			zap.Float64("delta_y", delta.DeltaY),
		)
		return
	}

	// Buffer the movement for processing in the next tick
	if _, exists := se.movementBuffer[entityID]; !exists {
		se.movementBuffer[entityID] = make([]*proto.MovementDelta, 0)
	}

	se.movementBuffer[entityID] = append(se.movementBuffer[entityID], delta)
}

func (se *SpatialEngine) processMovementDeltas() {
	for entityID, deltas := range se.movementBuffer {
		ent, exists := se.entityManager.GetEntity(entityID)
		if !exists {
			delete(se.movementBuffer, entityID)
			continue
		}

		// Process deltas in sequence order
		for _, delta := range deltas {
			if delta.Sequence > ent.LastSequence {
				// Apply movement validation
				newX := ent.Position.X + delta.DeltaX
				newY := ent.Position.Y + delta.DeltaY

				if se.isPositionValid(newX, newY) {
					// Update velocity based on movement delta
					ent.Velocity.X = delta.DeltaX
					ent.Velocity.Y = delta.DeltaY
					ent.LastSequence = delta.Sequence
				} else {
					// Movement would go out of bounds, generate correction
					se.generateCorrection(entityID)
				}
			}
		}

		// Clear processed deltas
		delete(se.movementBuffer, entityID)
	}
}

func (se *SpatialEngine) updateEntityPositions() {
	entities := se.entityManager.GetAllEntities()

	for _, ent := range entities {
		// Apply velocity to position
		newX := ent.Position.X + ent.Velocity.X
		newY := ent.Position.Y + ent.Velocity.Y

		// Apply friction
		ent.Velocity.X *= 0.95
		ent.Velocity.Y *= 0.95

		// Update position if valid
		if se.isPositionValid(newX, newY) {
			oldPos := ent.Position
			ent.Position.X = newX
			ent.Position.Y = newY

			// Update spatial index
			se.quadtree.Update(ent, oldPos)
		} else {
			// Clamp to bounds and stop velocity
			ent.Position.X = math.Max(se.config.WorldBounds.MinX, math.Min(se.config.WorldBounds.MaxX, newX))
			ent.Position.Y = math.Max(se.config.WorldBounds.MinY, math.Min(se.config.WorldBounds.MaxY, newY))
			ent.Velocity.X = 0
			ent.Velocity.Y = 0

			// Generate correction for out-of-bounds movement
			se.generateCorrection(ent.ID)
		}
	}
}

func (se *SpatialEngine) updateSpatialIndex() {
	// Spatial index is updated during entity position updates
	// This method can be used for additional spatial optimizations
}

func (se *SpatialEngine) processAOIEvents() {
	entities := se.entityManager.GetAllEntities()

	for _, ent := range entities {
		events := se.aoiManager.UpdateEntity(ent.ID, ent.Position)

		for _, event := range events {
			switch event.Type {
			case "enter", "move":
				// Send entity state to nearby entities
				se.broadcastToNearby(ent.ID, ent.Position)
			case "exit":
				// Send despawn notification to entities that are no longer nearby
				se.broadcastDespawnToNearby(ent.ID, ent.Position)
			}
		}
	}
}

func (se *SpatialEngine) broadcastToNearby(entityID uint32, position entity.Vector2) {
	nearbyEntities := se.aoiManager.GetEntitiesInRadius(position, se.config.AOIRadius)
	
	ent, exists := se.entityManager.GetEntity(entityID)
	if !exists {
		return
	}

	// Create entity state message
	entityState := &proto.EntityState{
		EntityId:    ent.ID,
		X:           float32(ent.Position.X),
		Y:           float32(ent.Position.Y),
		VelocityX:   float32(ent.Velocity.X),
		VelocityY:   float32(ent.Velocity.Y),
		LastUpdate:  uint64(time.Now().UnixMilli()),
		EntityType:   ent.Type,
	}

	for _, nearbyEnt := range nearbyEntities {
		if nearbyEnt.ID == entityID {
			continue // Don't send to self
		}

		// Queue broadcast for nearby entity
		message := &proto.Message{
			Type: proto.MessageType_ENTITY_STATE,
			Payload: &proto.Message_EntityState{
				EntityState: entityState,
			},
		}

		// This would be sent via the gateway
		se.queueBroadcast(nearbyEnt.ClientID, message)
	}
}

func (se *SpatialEngine) broadcastDespawnToNearby(entityID uint32, position entity.Vector2) {
	nearbyEntities := se.aoiManager.GetEntitiesInRadius(position, se.config.AOIRadius)

	for _, nearbyEnt := range nearbyEntities {
		if nearbyEnt.ID == entityID {
			continue
		}

		// Create despawn message
		message := &proto.Message{
			Type: proto.MessageType_DESPAWN,
			Payload: &proto.Message_Despawn{
				Despawn: &proto.Despawn{
					EntityId: entityID,
					Reason:   "out_of_aoi",
				},
			},
		}

		se.queueBroadcast(nearbyEnt.ClientID, message)
	}
}

func (se *SpatialEngine) generateCorrection(entityID uint32) {
	ent, exists := se.entityManager.GetEntity(entityID)
	if !exists {
		return
	}

	// Create correction message
	message := &proto.Message{
		Type: proto.MessageType_CORRECTION,
		Payload: &proto.Message_Correction{
			Correction: &proto.Correction{
				EntityId:        entityID,
				CorrectX:        float32(ent.Position.X),
				CorrectY:        float32(ent.Position.Y),
				CorrectVelocityX: float32(ent.Velocity.X),
				CorrectVelocityY: float32(ent.Velocity.Y),
				AckSequence:     ent.LastSequence,
			},
		},
	}

	se.queueBroadcast(ent.ClientID, message)
}

func (se *SpatialEngine) queueBroadcast(clientID string, message *proto.Message) {
	// This would be handled by the gateway
	// For now, we'll just log it
	se.logger.Debug("Queuing broadcast",
		zap.String("client_id", clientID),
		zap.String("message_type", message.Type.String()),
	)
}

func (se *SpatialEngine) broadcastWorker() {
	defer se.wg.Done()

	for {
		select {
		case <-se.shutdown:
			return
		case msg := <-se.broadcastChan:
			// This would send to the gateway
			se.logger.Debug("Broadcasting message",
				zap.String("client_id", msg.ClientID),
				zap.Int("data_size", len(msg.Data)),
			)
		}
	}
}

func (se *SpatialEngine) validateMovement(entityID uint32, delta *proto.MovementDelta) bool {
	ent, exists := se.entityManager.GetEntity(entityID)
	if !exists {
		return false
	}

	// Check sequence number
	if delta.Sequence <= ent.LastSequence {
		return false
	}

	// Check movement speed
	distance := math.Sqrt(delta.DeltaX*delta.DeltaX + delta.DeltaY*delta.DeltaY)
	if distance > se.config.MaxSpeed {
		return false
	}

	// Check if new position would be valid
	newX := ent.Position.X + delta.DeltaX
	newY := ent.Position.Y + delta.DeltaY

	return se.isPositionValid(newX, newY)
}

func (se *SpatialEngine) isPositionValid(x, y float64) bool {
	return x >= se.config.WorldBounds.MinX &&
		x <= se.config.WorldBounds.MaxX &&
		y >= se.config.WorldBounds.MinY &&
		y <= se.config.WorldBounds.MaxY
}

func (se *SpatialEngine) GetStats() map[string]interface{} {
	se.mu.RLock()
	defer se.mu.RUnlock()

	return map[string]interface{}{
		"entity_count":       se.entityManager.GetEntityCount(),
		"current_tick":       se.tickManager.GetCurrentTick(),
		"quadtree_stats":     se.quadtree.GetStats(),
		"aoi_subscribers":    se.aoiManager.GetSubscriberCount(),
		"movement_buffer_size": len(se.movementBuffer),
	}
}
