package engine

import (
	"math"
	"time"

	"github.com/akarsh-2004/aether/internal/config"
	"github.com/akarsh-2004/aether/internal/engine/entity"
	"github.com/akarsh-2004/aether/proto"
	"go.uber.org/zap"
)

type AuthoritySystem struct {
	config     config.EngineConfig
	logger     *zap.Logger
	validator  *MovementValidator
	reconciler *StateReconciler
}

type MovementValidator struct {
	config      config.EngineConfig
	logger      *zap.Logger
	speedHistory map[uint32]*SpeedHistory
}

type SpeedHistory struct {
	samples    []float64
	lastUpdate time.Time
}

type StateReconciler struct {
	config     config.EngineConfig
	logger     *zap.Logger
	corrections map[uint32]*CorrectionState
}

type CorrectionState struct {
	lastCorrection time.Time
	correctionCount int
	snapbackThreshold float64
}

func NewAuthoritySystem(cfg config.EngineConfig, logger *zap.Logger) *AuthoritySystem {
	return &AuthoritySystem{
		config:     cfg,
		logger:     logger,
		validator:  NewMovementValidator(cfg, logger),
		reconciler: NewStateReconciler(cfg, logger),
	}
}

func NewMovementValidator(cfg config.EngineConfig, logger *zap.Logger) *MovementValidator {
	return &MovementValidator{
		config:       cfg,
		logger:       logger,
		speedHistory: make(map[uint32]*SpeedHistory),
	}
}

func NewStateReconciler(cfg config.EngineConfig, logger *zap.Logger) *StateReconciler {
	return &StateReconciler{
		config:      cfg,
		logger:      logger,
		corrections: make(map[uint32]*CorrectionState),
	}
}

func (as *AuthoritySystem) ValidateMovement(ent *entity.Entity, delta *proto.MovementDelta) ValidationResult {
	return as.validator.Validate(ent, delta)
}

func (as *AuthoritySystem) ShouldReconcile(ent *entity.Entity, clientState *proto.EntityState) bool {
	return as.reconciler.ShouldReconcile(ent, clientState)
}

func (as *AuthoritySystem) GenerateCorrection(ent *entity.Entity, ackSequence uint64) *proto.Correction {
	return as.reconciler.GenerateCorrection(ent, ackSequence)
}

type ValidationResult struct {
	Valid    bool
	Reason   string
	Modified bool
	NewDelta *proto.MovementDelta
}

func (mv *MovementValidator) Validate(ent *entity.Entity, delta *proto.MovementDelta) ValidationResult {
	result := ValidationResult{Valid: true}

	// Check sequence number
	if delta.Sequence <= ent.LastSequence {
		result.Valid = false
		result.Reason = "outdated sequence"
		return result
	}

	// Check movement speed
	distance := math.Sqrt(delta.DeltaX*delta.DeltaX + delta.DeltaY*delta.DeltaY)
	if distance > mv.config.MaxSpeed {
		result.Valid = false
		result.Reason = "exceeds max speed"
		result.Modified = true
		result.NewDelta = mv.limitSpeed(delta, mv.config.MaxSpeed)
		return result
	}

	// Check for teleportation (large position jumps)
	if mv.isTeleportation(ent, delta) {
		result.Valid = false
		result.Reason = "teleportation detected"
		return result
	}

	// Update speed history for anomaly detection
	mv.updateSpeedHistory(ent.ID, distance)

	// Check for speed anomalies
	if mv.hasSpeedAnomaly(ent.ID) {
		mv.logger.Warn("Speed anomaly detected",
			zap.Uint32("entity_id", ent.ID),
			zap.Float64("distance", distance),
			zap.Float64("max_speed", mv.config.MaxSpeed),
		)
		
		// Allow movement but flag for monitoring
	}

	// Check bounds
	newX := ent.Position.X + delta.DeltaX
	newY := ent.Position.Y + delta.DeltaY

	if !mv.isInBounds(newX, newY) {
		result.Valid = false
		result.Reason = "out of bounds"
		result.Modified = true
		result.NewDelta = mv.clampToBounds(ent, delta)
		return result
	}

	return result
}

func (mv *MovementValidator) limitSpeed(delta *proto.MovementDelta, maxSpeed float64) *proto.MovementDelta {
	distance := math.Sqrt(delta.DeltaX*delta.DeltaX + delta.DeltaY*delta.DeltaY)
	if distance <= maxSpeed {
		return delta
	}

	scale := maxSpeed / distance
	return &proto.MovementDelta{
		EntityId:  delta.EntityId,
		Sequence:  delta.Sequence,
		DeltaX:    delta.DeltaX * scale,
		DeltaY:    delta.DeltaY * scale,
		Timestamp: delta.Timestamp,
	}
}

func (mv *MovementValidator) isTeleportation(ent *entity.Entity, delta *proto.MovementDelta) bool {
	// Define teleportation threshold (e.g., 3x max speed)
	teleportThreshold := mv.config.MaxSpeed * 3
	
	distance := math.Sqrt(delta.DeltaX*delta.DeltaX + delta.DeltaY*delta.DeltaY)
	return distance > teleportThreshold
}

func (mv *MovementValidator) isInBounds(x, y float64) bool {
	return x >= mv.config.WorldBounds.MinX &&
		x <= mv.config.WorldBounds.MaxX &&
		y >= mv.config.WorldBounds.MinY &&
		y <= mv.config.WorldBounds.MaxY
}

func (mv *MovementValidator) clampToBounds(ent *entity.Entity, delta *proto.MovementDelta) *proto.MovementDelta {
	newX := ent.Position.X + delta.DeltaX
	newY := ent.Position.Y + delta.DeltaY

	// Clamp to bounds
	clampedX := math.Max(mv.config.WorldBounds.MinX, math.Min(mv.config.WorldBounds.MaxX, newX))
	clampedY := math.Max(mv.config.WorldBounds.MinY, math.Min(mv.config.WorldBounds.MaxY, newY))

	// Calculate new delta
	clampedDeltaX := clampedX - ent.Position.X
	clampedDeltaY := clampedY - ent.Position.Y

	return &proto.MovementDelta{
		EntityId:  delta.EntityId,
		Sequence:  delta.Sequence,
		DeltaX:    clampedDeltaX,
		DeltaY:    clampedDeltaY,
		Timestamp: delta.Timestamp,
	}
}

func (mv *MovementValidator) updateSpeedHistory(entityID uint32, speed float64) {
	history, exists := mv.speedHistory[entityID]
	if !exists {
		history = &SpeedHistory{
			samples:    make([]float64, 0, 10),
			lastUpdate: time.Now(),
		}
		mv.speedHistory[entityID] = history
	}

	// Add new sample
	history.samples = append(history.samples, speed)
	
	// Keep only last 10 samples
	if len(history.samples) > 10 {
		history.samples = history.samples[1:]
	}

	history.lastUpdate = time.Now()
}

func (mv *MovementValidator) hasSpeedAnomaly(entityID uint32) bool {
	history, exists := mv.speedHistory[entityID]
	if !exists || len(history.samples) < 5 {
		return false
	}

	// Calculate average speed
	var sum float64
	for _, speed := range history.samples {
		sum += speed
	}
	avgSpeed := sum / float64(len(history.samples))

	// Check if recent speeds are significantly higher than average
	recentSamples := history.samples[len(history.samples)-3:]
	for _, speed := range recentSamples {
		if speed > avgSpeed*2.5 { // 250% of average
			return true
		}
	}

	return false
}

func (sr *StateReconciler) ShouldReconcile(ent *entity.Entity, clientState *proto.EntityState) bool {
	// Calculate position difference
	serverX := ent.Position.X
	serverY := ent.Position.Y
	clientX := float64(clientState.X)
	clientY := float64(clientState.Y)

	distance := math.Sqrt((serverX-clientX)*(serverX-clientX) + (serverY-clientY)*(serverY-clientY))

	// Define reconciliation threshold
	reconciliationThreshold := sr.config.MaxSpeed * 2 // Allow some prediction error

	return distance > reconciliationThreshold
}

func (sr *StateReconciler) GenerateCorrection(ent *entity.Entity, ackSequence uint64) *proto.Correction {
	correction := &proto.Correction{
		EntityId:         ent.ID,
		CorrectX:         float32(ent.Position.X),
		CorrectY:         float32(ent.Position.Y),
		CorrectVelocityX: float32(ent.Velocity.X),
		CorrectVelocityY: float32(ent.Velocity.Y),
		AckSequence:      ackSequence,
	}

	// Track correction for rate limiting
	sr.trackCorrection(ent.ID)

	return correction
}

func (sr *StateReconciler) trackCorrection(entityID uint32) {
	state, exists := sr.corrections[entityID]
	if !exists {
		state = &CorrectionState{
			lastCorrection:     time.Now(),
			correctionCount:    0,
			snapbackThreshold:  sr.config.MaxSpeed * 5,
		}
		sr.corrections[entityID] = state
	}

	state.lastCorrection = time.Now()
	state.correctionCount++

	// Log frequent corrections
	if state.correctionCount%10 == 0 {
		sr.logger.Warn("Frequent corrections for entity",
			zap.Uint32("entity_id", entityID),
			zap.Int("correction_count", state.correctionCount),
			zap.Time("last_correction", state.lastCorrection),
		)
	}

	// Clean up old correction states
	sr.cleanupOldCorrectionStates()
}

func (sr *StateReconciler) cleanupOldCorrectionStates() {
	// Clean up correction states older than 5 minutes
	cutoff := time.Now().Add(-5 * time.Minute)

	for entityID, state := range sr.corrections {
		if state.lastUpdate.Before(cutoff) {
			delete(sr.corrections, entityID)
		}
	}
}

func (sr *StateReconciler) GetCorrectionStats(entityID uint32) (int, time.Time) {
	state, exists := sr.corrections[entityID]
	if !exists {
		return 0, time.Time{}
	}

	return state.correctionCount, state.lastCorrection
}
