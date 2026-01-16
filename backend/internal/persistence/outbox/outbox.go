package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/akarsh-2004/aether/internal/persistence/postgres"
	"go.uber.org/zap"
)

type OutboxProcessor struct {
	pgClient   *postgres.PostgresClient
	logger     *zap.Logger
	handlers   map[string]EventHandler
	buffer     []OutboxEvent
	bufferSize int
	mu         sync.RWMutex
	running    bool
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

type OutboxEvent struct {
	ID        int64                  `json:"id"`
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
}

type EventHandler func(ctx context.Context, payload map[string]interface{}) error

func NewOutboxProcessor(pgClient *postgres.PostgresClient, logger *zap.Logger) *OutboxProcessor {
	return &OutboxProcessor{
		pgClient:   pgClient,
		logger:     logger,
		handlers:   make(map[string]EventHandler),
		bufferSize: 100,
		stopChan:   make(chan struct{}),
	}
}

func (op *OutboxProcessor) RegisterHandler(eventType string, handler EventHandler) {
	op.mu.Lock()
	defer op.mu.Unlock()
	
	op.handlers[eventType] = handler
	op.logger.Info("Registered outbox event handler", zap.String("event_type", eventType))
}

func (op *OutboxProcessor) Start(ctx context.Context) error {
	op.mu.Lock()
	if op.running {
		op.mu.Unlock()
		return fmt.Errorf("outbox processor is already running")
	}
	op.running = true
	op.mu.Unlock()

	op.wg.Add(1)
	go op.processLoop(ctx)

	op.logger.Info("Outbox processor started")
	return nil
}

func (op *OutboxProcessor) Stop() {
	op.mu.Lock()
	if !op.running {
		op.mu.Unlock()
		return
	}
	op.running = false
	op.mu.Unlock()

	close(op.stopChan)
	op.wg.Wait()

	op.logger.Info("Outbox processor stopped")
}

func (op *OutboxProcessor) processLoop(ctx context.Context) {
	defer op.wg.Done()

	ticker := time.NewTicker(1 * time.Second) // Process every second
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-op.stopChan:
			return
		case <-ticker.C:
			op.processBatch(ctx)
		}
	}
}

func (op *OutboxProcessor) processBatch(ctx context.Context) {
	// Get unprocessed events
	events, err := op.pgClient.GetUnprocessedOutboxEvents(ctx, op.bufferSize)
	if err != nil {
		op.logger.Error("Failed to get unprocessed outbox events", zap.Error(err))
		return
	}

	if len(events) == 0 {
		return
	}

	op.logger.Debug("Processing outbox batch", zap.Int("event_count", len(events)))

	for _, event := range events {
		if err := op.processEvent(ctx, event); err != nil {
			op.logger.Error("Failed to process outbox event",
				zap.Int64("event_id", event.ID),
				zap.String("event_type", event.EventType),
				zap.Error(err),
			)
			continue
		}

		// Mark as processed
		if err := op.pgClient.MarkOutboxEventProcessed(ctx, event.ID); err != nil {
			op.logger.Error("Failed to mark outbox event as processed",
				zap.Int64("event_id", event.ID),
				zap.Error(err),
			)
		}
	}
}

func (op *OutboxProcessor) processEvent(ctx context.Context, pgEvent *postgres.OutboxEvent) error {
	// Parse payload
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(pgEvent.Payload), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Get handler
	op.mu.RLock()
	handler, exists := op.handlers[pgEvent.EventType]
	op.mu.RUnlock()

	if !exists {
		op.logger.Warn("No handler registered for event type",
			zap.String("event_type", pgEvent.EventType),
		)
		return nil // Not an error, just no handler
	}

	// Execute handler
	if err := handler(ctx, payload); err != nil {
		return fmt.Errorf("handler failed: %w", err)
	}

	op.logger.Debug("Processed outbox event",
		zap.Int64("event_id", pgEvent.ID),
		zap.String("event_type", pgEvent.EventType),
	)

	return nil
}

func (op *OutboxProcessor) PublishEvent(ctx context.Context, eventType string, payload map[string]interface{}) error {
	return op.pgClient.SaveOutboxEvent(ctx, eventType, payload)
}

// Built-in event handlers

func (op *OutboxProcessor) RegisterBuiltInHandlers() {
	// Entity spawn event handler
	op.RegisterHandler("entity_spawned", func(ctx context.Context, payload map[string]interface{}) error {
		entityID, ok := payload["entity_id"].(float64)
		if !ok {
			return fmt.Errorf("invalid entity_id in payload")
		}

		clientID, ok := payload["client_id"].(string)
		if !ok {
			return fmt.Errorf("invalid client_id in payload")
		}

		op.logger.Info("Entity spawned event processed",
			zap.Uint32("entity_id", uint32(entityID)),
			zap.String("client_id", clientID),
		)

		return nil
	})

	// Entity despawn event handler
	op.RegisterHandler("entity_despawned", func(ctx context.Context, payload map[string]interface{}) error {
		entityID, ok := payload["entity_id"].(float64)
		if !ok {
			return fmt.Errorf("invalid entity_id in payload")
		}

		reason, ok := payload["reason"].(string)
		if !ok {
			reason = "unknown"
		}

		op.logger.Info("Entity despawned event processed",
			zap.Uint32("entity_id", uint32(entityID)),
			zap.String("reason", reason),
		)

		return nil
	})

	// Movement correction event handler
	op.RegisterHandler("movement_correction", func(ctx context.Context, payload map[string]interface{}) error {
		entityID, ok := payload["entity_id"].(float64)
		if !ok {
			return fmt.Errorf("invalid entity_id in payload")
		}

		clientID, ok := payload["client_id"].(string)
		if !ok {
			return fmt.Errorf("invalid client_id in payload")
		}

		op.logger.Info("Movement correction event processed",
			zap.Uint32("entity_id", uint32(entityID)),
			zap.String("client_id", clientID),
		)

		return nil
	})

	// Session started event handler
	op.RegisterHandler("session_started", func(ctx context.Context, payload map[string]interface{}) error {
		sessionID, ok := payload["session_id"].(string)
		if !ok {
			return fmt.Errorf("invalid session_id in payload")
		}

		clientID, ok := payload["client_id"].(string)
		if !ok {
			return fmt.Errorf("invalid client_id in payload")
		}

		op.logger.Info("Session started event processed",
			zap.String("session_id", sessionID),
			zap.String("client_id", clientID),
		)

		return nil
	})

	// Session ended event handler
	op.RegisterHandler("session_ended", func(ctx context.Context, payload map[string]interface{}) error {
		sessionID, ok := payload["session_id"].(string)
		if !ok {
			return fmt.Errorf("invalid session_id in payload")
		}

		duration, ok := payload["duration_seconds"].(float64)
		if !ok {
			duration = 0
		}

		op.logger.Info("Session ended event processed",
			zap.String("session_id", sessionID),
			zap.Float64("duration_seconds", duration),
		)

		return nil
	})
}

func (op *OutboxProcessor) GetStats() map[string]interface{} {
	op.mu.RLock()
	defer op.mu.RUnlock()

	return map[string]interface{}{
		"running":       op.running,
		"handlers":      len(op.handlers),
		"buffer_size":   op.bufferSize,
		"registered_events": func() []string {
			events := make([]string, 0, len(op.handlers))
			for eventType := range op.handlers {
				events = append(events, eventType)
			}
			return events
		}(),
	}
}
