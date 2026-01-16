package tick

import (
	"context"
	"sync"
	"time"

	"github.com/akarsh-2004/aether/internal/config"
	"go.uber.org/zap"
)

type TickManager struct {
	config       config.EngineConfig
	logger       *zap.Logger
	tickRate     time.Duration
	currentTick  uint64
	shutdown     chan struct{}
	wg           sync.WaitGroup
	tickHandlers []TickHandler
	mu           sync.RWMutex
}

type TickHandler interface {
	OnTick(tickNumber uint64)
	OnShutdown()
}

type TickContext struct {
	TickNumber uint64
	DeltaTime  time.Duration
}

func NewTickManager(cfg config.EngineConfig, logger *zap.Logger) *TickManager {
	return &TickManager{
		config:      cfg,
		logger:      logger,
		tickRate:    time.Duration(cfg.TickRateMs) * time.Millisecond,
		currentTick: 0,
		shutdown:    make(chan struct{}),
	}
}

func (tm *TickManager) Start(ctx context.Context) error {
	ticker := time.NewTicker(tm.tickRate)
	defer ticker.Stop()

	tm.wg.Add(1)
	defer tm.wg.Done()

	tm.logger.Info("Tick loop started",
		zap.Duration("tick_rate", tm.tickRate),
		zap.Int("target_hz", 1000/tm.config.TickRateMs),
	)

	for {
		select {
		case <-ctx.Done():
			tm.logger.Info("Tick loop shutting down")
			tm.shutdownHandlers()
			return ctx.Err()
		case <-tm.shutdown:
			tm.logger.Info("Tick loop shutting down")
			tm.shutdownHandlers()
			return nil
		case <-ticker.C:
			tm.processTick()
		}
	}
}

func (tm *TickManager) Shutdown() {
	close(tm.shutdown)
	tm.wg.Wait()
}

func (tm *TickManager) AddHandler(handler TickHandler) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.tickHandlers = append(tm.tickHandlers, handler)
}

func (tm *TickManager) GetCurrentTick() uint64 {
	return tm.currentTick
}

func (tm *TickManager) processTick() {
	start := time.Now()
	tm.currentTick++

	tm.mu.RLock()
	handlers := make([]TickHandler, len(tm.tickHandlers))
	copy(handlers, tm.tickHandlers)
	tm.mu.RUnlock()

	// Execute all handlers for this tick
	for _, handler := range handlers {
		handler.OnTick(tm.currentTick)
	}

	duration := time.Since(start)
	
	// Log warning if tick processing takes too long
	if duration > tm.tickRate/2 {
		tm.logger.Warn("Tick processing taking too long",
			zap.Uint64("tick", tm.currentTick),
			zap.Duration("duration", duration),
			zap.Duration("tick_rate", tm.tickRate),
		)
	}

	// Debug logging every 100 ticks
	if tm.currentTick%100 == 0 {
		tm.logger.Debug("Tick processed",
			zap.Uint64("tick", tm.currentTick),
			zap.Duration("duration", duration),
		)
	}
}

func (tm *TickManager) shutdownHandlers() {
	tm.mu.RLock()
	handlers := make([]TickHandler, len(tm.tickHandlers))
	copy(handlers, tm.tickHandlers)
	tm.mu.RUnlock()

	for _, handler := range handlers {
		handler.OnShutdown()
	}
}
