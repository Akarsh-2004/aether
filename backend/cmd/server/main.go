package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akarsh-2004/aether/internal/config"
	"github.com/akarsh-2004/aether/internal/engine"
	"github.com/akarsh-2004/aether/internal/gateway"
	"github.com/akarsh-2004/aether/internal/observability"
	"go.uber.org/zap"
)

func main() {
	var configPath = flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	logger, err := observability.NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spatialEngine := engine.NewSpatialEngine(cfg.Engine, logger)
	if err := spatialEngine.Start(ctx); err != nil {
		logger.Fatal("Failed to start spatial engine", zap.Error(err))
	}

	wsGateway := gateway.NewWebSocketGateway(cfg.Gateway, spatialEngine, logger)
	if err := wsGateway.Start(ctx); err != nil {
		logger.Fatal("Failed to start WebSocket gateway", zap.Error(err))
	}

	logger.Info("Aether server started successfully",
		zap.String("version", "1.0.0"),
		zap.String("bind_addr", cfg.Gateway.BindAddr),
		zap.Int("tick_rate_ms", cfg.Engine.TickRateMs),
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutdown signal received, initiating graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := wsGateway.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error during WebSocket gateway shutdown", zap.Error(err))
	}

	if err := spatialEngine.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error during spatial engine shutdown", zap.Error(err))
	}

	logger.Info("Aether server shutdown complete")
}
