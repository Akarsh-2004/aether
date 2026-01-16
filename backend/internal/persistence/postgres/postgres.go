package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/akarsh-2004/aether/internal/config"
	"go.uber.org/zap"
)

type PostgresClient struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	config config.PostgresConfig
}

type EntitySnapshot struct {
	ID           uint32    `json:"id"`
	Type         string    `json:"type"`
	PositionX    float64   `json:"position_x"`
	PositionY    float64   `json:"position_y"`
	VelocityX    float64   `json:"velocity_x"`
	VelocityY    float64   `json:"velocity_y"`
	ClientID     string    `json:"client_id"`
	LastUpdate   time.Time `json:"last_update"`
	TickNumber   uint64    `json:"tick_number"`
}

type OutboxEvent struct {
	ID        int64     `json:"id"`
	EventType string    `json:"event_type"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
	Processed bool      `json:"processed"`
}

func NewPostgresClient(cfg config.PostgresConfig, logger *zap.Logger) (*PostgresClient, error) {
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	config.MaxConns = int32(cfg.MaxOpenConns)
	config.MinConns = int32(cfg.MaxIdleConns)
	config.MaxConnLifetime = time.Duration(cfg.ConnMaxLifetime) * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	client := &PostgresClient{
		pool:   pool,
		logger: logger,
		config: cfg,
	}

	// Initialize schema
	if err := client.initializeSchema(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return client, nil
}

func (p *PostgresClient) Close() {
	p.pool.Close()
}

func (p *PostgresClient) initializeSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS entity_snapshots (
			id SERIAL PRIMARY KEY,
			entity_id INTEGER NOT NULL,
			entity_type VARCHAR(50) NOT NULL,
			position_x DOUBLE PRECISION NOT NULL,
			position_y DOUBLE PRECISION NOT NULL,
			velocity_x DOUBLE PRECISION NOT NULL,
			velocity_y DOUBLE PRECISION NOT NULL,
			client_id VARCHAR(100) NOT NULL,
			last_update TIMESTAMP WITH TIME ZONE NOT NULL,
			tick_number BIGINT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		
		`CREATE INDEX IF NOT EXISTS idx_entity_snapshots_entity_id ON entity_snapshots(entity_id)`,
		`CREATE INDEX IF NOT EXISTS idx_entity_snapshots_client_id ON entity_snapshots(client_id)`,
		`CREATE INDEX IF NOT EXISTS idx_entity_snapshots_created_at ON entity_snapshots(created_at)`,
		
		`CREATE TABLE IF NOT EXISTS outbox_events (
			id SERIAL PRIMARY KEY,
			event_type VARCHAR(50) NOT NULL,
			payload JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			processed BOOLEAN DEFAULT FALSE,
			processed_at TIMESTAMP WITH TIME ZONE
		)`,
		
		`CREATE INDEX IF NOT EXISTS idx_outbox_events_processed ON outbox_events(processed, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_outbox_events_event_type ON outbox_events(event_type)`,
		
		`CREATE TABLE IF NOT EXISTS game_sessions (
			id SERIAL PRIMARY KEY,
			session_id VARCHAR(100) UNIQUE NOT NULL,
			client_id VARCHAR(100) NOT NULL,
			entity_id INTEGER,
			start_time TIMESTAMP WITH TIME ZONE NOT NULL,
			end_time TIMESTAMP WITH TIME ZONE,
			duration_seconds INTEGER,
			metadata JSONB
		)`,
		
		`CREATE INDEX IF NOT EXISTS idx_game_sessions_client_id ON game_sessions(client_id)`,
		`CREATE INDEX IF NOT EXISTS idx_game_sessions_start_time ON game_sessions(start_time)`,
	}

	for _, query := range queries {
		if _, err := p.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	p.logger.Info("Database schema initialized")
	return nil
}

func (p *PostgresClient) SaveEntitySnapshot(ctx context.Context, snapshot *EntitySnapshot) error {
	query := `
		INSERT INTO entity_snapshots 
		(entity_id, entity_type, position_x, position_y, velocity_x, velocity_y, client_id, last_update, tick_number)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := p.pool.Exec(ctx, query,
		snapshot.ID,
		snapshot.Type,
		snapshot.PositionX,
		snapshot.PositionY,
		snapshot.VelocityX,
		snapshot.VelocityY,
		snapshot.ClientID,
		snapshot.LastUpdate,
		snapshot.TickNumber,
	)

	if err != nil {
		return fmt.Errorf("failed to save entity snapshot: %w", err)
	}

	return nil
}

func (p *PostgresClient) GetLatestEntitySnapshot(ctx context.Context, entityID uint32) (*EntitySnapshot, error) {
	query := `
		SELECT entity_id, entity_type, position_x, position_y, velocity_x, velocity_y, client_id, last_update, tick_number
		FROM entity_snapshots
		WHERE entity_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var snapshot EntitySnapshot
	err := p.pool.QueryRow(ctx, query, entityID).Scan(
		&snapshot.ID,
		&snapshot.Type,
		&snapshot.PositionX,
		&snapshot.PositionY,
		&snapshot.VelocityX,
		&snapshot.VelocityY,
		&snapshot.ClientID,
		&snapshot.LastUpdate,
		&snapshot.TickNumber,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get entity snapshot: %w", err)
	}

	return &snapshot, nil
}

func (p *PostgresClient) SaveOutboxEvent(ctx context.Context, eventType string, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		INSERT INTO outbox_events (event_type, payload)
		VALUES ($1, $2)
	`

	_, err = p.pool.Exec(ctx, query, eventType, payloadJSON)
	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	return nil
}

func (p *PostgresClient) GetUnprocessedOutboxEvents(ctx context.Context, limit int) ([]*OutboxEvent, error) {
	query := `
		SELECT id, event_type, payload, created_at, processed
		FROM outbox_events
		WHERE processed = FALSE
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := p.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query outbox events: %w", err)
	}
	defer rows.Close()

	var events []*OutboxEvent
	for rows.Next() {
		var event OutboxEvent
		if err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.Payload,
			&event.CreatedAt,
			&event.Processed,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outbox event: %w", err)
		}
		events = append(events, &event)
	}

	return events, nil
}

func (p *PostgresClient) MarkOutboxEventProcessed(ctx context.Context, eventID int64) error {
	query := `
		UPDATE outbox_events
		SET processed = TRUE, processed_at = NOW()
		WHERE id = $1
	`

	_, err := p.pool.Exec(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to mark outbox event as processed: %w", err)
	}

	return nil
}

func (p *PostgresClient) StartGameSession(ctx context.Context, sessionID, clientID string, entityID *uint32) error {
	query := `
		INSERT INTO game_sessions (session_id, client_id, entity_id, start_time)
		VALUES ($1, $2, $3, NOW())
	`

	_, err := p.pool.Exec(ctx, query, sessionID, clientID, entityID)
	if err != nil {
		return fmt.Errorf("failed to start game session: %w", err)
	}

	return nil
}

func (p *PostgresClient) EndGameSession(ctx context.Context, sessionID string) error {
	query := `
		UPDATE game_sessions
		SET end_time = NOW(),
			duration_seconds = EXTRACT(EPOCH FROM (NOW() - start_time))::INTEGER
		WHERE session_id = $1 AND end_time IS NULL
	`

	result, err := p.pool.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to end game session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found or already ended")
	}

	return nil
}

func (p *PostgresClient) CleanupOldSnapshots(ctx context.Context, olderThan time.Duration) error {
	query := `
		DELETE FROM entity_snapshots
		WHERE created_at < NOW() - INTERVAL '1 second' * $1
	`

	result, err := p.pool.Exec(ctx, query, int64(olderThan.Seconds()))
	if err != nil {
		return fmt.Errorf("failed to cleanup old snapshots: %w", err)
	}

	p.logger.Info("Cleaned up old snapshots", zap.Int64("deleted_count", result.RowsAffected()))
	return nil
}

func (p *PostgresClient) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get connection pool stats
	stats := make(map[string]interface{})
	poolStats := p.pool.Stat()
	stats["total_connections"] = poolStats.TotalConns()
	stats["idle_connections"] = poolStats.IdleConns()
	stats["acquired_connections"] = poolStats.AcquiredConns()

	// Get table counts
	tables := []string{"entity_snapshots", "outbox_events", "game_sessions"}
	for _, table := range tables {
		var count int64
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		if err := p.pool.QueryRow(ctx, query).Scan(&count); err != nil {
			p.logger.Warn("Failed to get table count", zap.String("table", table), zap.Error(err))
		} else {
			stats[fmt.Sprintf("%s_count", table)] = count
		}
	}

	return stats, nil
}
