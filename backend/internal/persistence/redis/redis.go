package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/akarsh-2004/aether/internal/config"
	"go.uber.org/zap"
)

type RedisClient struct {
	client *redis.Client
	logger *zap.Logger
	config config.RedisConfig
}

type PresenceData struct {
	ClientID    string    `json:"client_id"`
	EntityID    uint32    `json:"entity_id"`
	LastSeen    time.Time `json:"last_seen"`
	PositionX   float64   `json:"position_x"`
	PositionY   float64   `json:"position_y"`
}

func NewRedisClient(cfg config.RedisConfig, logger *zap.Logger) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		client: rdb,
		logger: logger,
		config: cfg,
	}, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) UpdatePresence(ctx context.Context, clientID string, entityID uint32, x, y float64) error {
	presence := PresenceData{
		ClientID:  clientID,
		EntityID:  entityID,
		LastSeen:  time.Now(),
		PositionX: x,
		PositionY: y,
	}

	data, err := json.Marshal(presence)
	if err != nil {
		return fmt.Errorf("failed to marshal presence data: %w", err)
	}

	key := fmt.Sprintf("presence:%s", clientID)
	
	// Set with TTL of 5 minutes
	if err := r.client.Set(ctx, key, data, 5*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to set presence: %w", err)
	}

	// Also add to active sessions set
	sessionKey := "active_sessions"
	if err := r.client.SAdd(ctx, sessionKey, clientID).Err(); err != nil {
		r.logger.Warn("Failed to add to active sessions", zap.Error(err))
	}

	// Set TTL on sessions set
	r.client.Expire(ctx, sessionKey, 10*time.Minute)

	return nil
}

func (r *RedisClient) GetPresence(ctx context.Context, clientID string) (*PresenceData, error) {
	key := fmt.Sprintf("presence:%s", clientID)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get presence: %w", err)
	}

	var presence PresenceData
	if err := json.Unmarshal([]byte(data), &presence); err != nil {
		return nil, fmt.Errorf("failed to unmarshal presence data: %w", err)
	}

	return &presence, nil
}

func (r *RedisClient) RemovePresence(ctx context.Context, clientID string) error {
	key := fmt.Sprintf("presence:%s", clientID)
	
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to remove presence: %w", err)
	}

	// Remove from active sessions
	sessionKey := "active_sessions"
	if err := r.client.SRem(ctx, sessionKey, clientID).Err(); err != nil {
		r.logger.Warn("Failed to remove from active sessions", zap.Error(err))
	}

	return nil
}

func (r *RedisClient) GetActiveSessions(ctx context.Context) ([]string, error) {
	sessionKey := "active_sessions"
	members, err := r.client.SMembers(ctx, sessionKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	return members, nil
}

func (r *RedisClient) UpdateHeartbeat(ctx context.Context, clientID string) error {
	key := fmt.Sprintf("heartbeat:%s", clientID)
	
	// Set heartbeat with TTL of 30 seconds
	if err := r.client.Set(ctx, key, time.Now().Unix(), 30*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	return nil
}

func (r *RedisClient) IsClientAlive(ctx context.Context, clientID string) (bool, error) {
	key := fmt.Sprintf("heartbeat:%s", clientID)
	
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check heartbeat: %w", err)
	}

	return exists > 0, nil
}

func (r *RedisClient) CleanupStaleSessions(ctx context.Context) error {
	// Get all active sessions
	sessions, err := r.GetActiveSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active sessions: %w", err)
	}

	var staleSessions []string
	for _, clientID := range sessions {
		alive, err := r.IsClientAlive(ctx, clientID)
		if err != nil {
			r.logger.Warn("Failed to check client aliveness", zap.String("client_id", clientID), zap.Error(err))
			continue
		}

		if !alive {
			staleSessions = append(staleSessions, clientID)
		}
	}

	// Remove stale sessions
	for _, clientID := range staleSessions {
		if err := r.RemovePresence(ctx, clientID); err != nil {
			r.logger.Warn("Failed to remove stale presence", zap.String("client_id", clientID), zap.Error(err))
		} else {
			r.logger.Info("Cleaned up stale session", zap.String("client_id", clientID))
		}
	}

	r.logger.Info("Cleanup completed", zap.Int("stale_sessions_removed", len(staleSessions)))
	return nil
}

func (r *RedisClient) SetSessionData(ctx context.Context, clientID string, key string, value interface{}) error {
	fullKey := fmt.Sprintf("session:%s:%s", clientID, key)
	
	if err := r.client.Set(ctx, fullKey, value, 10*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to set session data: %w", err)
	}

	return nil
}

func (r *RedisClient) GetSessionData(ctx context.Context, clientID string, key string) (string, error) {
	fullKey := fmt.Sprintf("session:%s:%s", clientID, key)
	
	value, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Not found
		}
		return "", fmt.Errorf("failed to get session data: %w", err)
	}

	return value, nil
}

func (r *RedisClient) IncrementCounter(ctx context.Context, key string) (int64, error) {
	result, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment counter: %w", err)
	}

	// Set TTL if this is a new counter
	if result == 1 {
		r.client.Expire(ctx, key, time.Hour)
	}

	return result, nil
}

func (r *RedisClient) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := r.client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Parse basic stats
	stats := make(map[string]interface{})
	stats["info"] = info
	
	// Get active sessions count
	sessions, err := r.GetActiveSessions(ctx)
	if err != nil {
		r.logger.Warn("Failed to get active sessions for stats", zap.Error(err))
	} else {
		stats["active_sessions"] = len(sessions)
	}

	return stats, nil
}
