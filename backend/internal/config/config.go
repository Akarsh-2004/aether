package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Engine  EngineConfig  `yaml:"engine"`
	Gateway GatewayConfig `yaml:"gateway"`
	Redis   RedisConfig   `yaml:"redis"`
	Postgres PostgresConfig `yaml:"postgres"`
}

type EngineConfig struct {
	TickRateMs    int     `yaml:"tick_rate_ms"`    // Fixed timestep: 20-30ms
	MaxEntities   int     `yaml:"max_entities"`    // Target: 100-300 concurrent entities
	WorldBounds   Bounds  `yaml:"world_bounds"`    // World boundaries
	MaxSpeed      float64 `yaml:"max_speed"`       // Max movement speed per tick
	AOIRadius     float64 `yaml:"aoi_radius"`      // Area of Interest radius
	QuadtreeDepth int     `yaml:"quadtree_depth"`  // Maximum quadtree depth
	QuadtreeCapacity int  `yaml:"quadtree_capacity"` // Entities per quadtree node
}

type GatewayConfig struct {
	BindAddr         string `yaml:"bind_addr"`          // WebSocket bind address
	ReadBufferSize   int    `yaml:"read_buffer_size"`   // Read buffer size in bytes
	WriteBufferSize  int    `yaml:"write_buffer_size"`  // Write buffer size in bytes
	PingPeriod       int    `yaml:"ping_period"`        // Ping period in seconds
	PongWait         int    `yaml:"pong_wait"`          // Pong wait timeout in seconds
	WriteWait        int    `yaml:"write_wait"`         // Write wait timeout in seconds
	MaxMessageSize   int64  `yaml:"max_message_size"`   // Maximum message size
	EnableCompression bool  `yaml:"enable_compression"` // Enable WebSocket compression
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`      // Redis server address
	Password string `yaml:"password"`  // Redis password
	DB       int    `yaml:"db"`        // Redis database number
	PoolSize int    `yaml:"pool_size"` // Connection pool size
}

type PostgresConfig struct {
	Host            string `yaml:"host"`             // PostgreSQL host
	Port            int    `yaml:"port"`             // PostgreSQL port
	User            string `yaml:"user"`             // PostgreSQL user
	Password        string `yaml:"password"`         // PostgreSQL password
	DBName          string `yaml:"dbname"`           // PostgreSQL database name
	MaxOpenConns    int    `yaml:"max_open_conns"`    // Maximum open connections
	MaxIdleConns    int    `yaml:"max_idle_conns"`    // Maximum idle connections
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // Connection max lifetime in seconds
}

type Bounds struct {
	MinX float64 `yaml:"min_x"`
	MinY float64 `yaml:"min_y"`
	MaxX float64 `yaml:"max_x"`
	MaxY float64 `yaml:"max_y"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Engine.TickRateMs < 10 || c.Engine.TickRateMs > 100 {
		return fmt.Errorf("engine.tick_rate_ms must be between 10-100ms, got %d", c.Engine.TickRateMs)
	}

	if c.Engine.MaxEntities < 10 || c.Engine.MaxEntities > 1000 {
		return fmt.Errorf("engine.max_entities must be between 10-1000, got %d", c.Engine.MaxEntities)
	}

	if c.Engine.MaxSpeed <= 0 {
		return fmt.Errorf("engine.max_speed must be positive, got %f", c.Engine.MaxSpeed)
	}

	if c.Engine.AOIRadius <= 0 {
		return fmt.Errorf("engine.aoi_radius must be positive, got %f", c.Engine.AOIRadius)
	}

	if c.Gateway.BindAddr == "" {
		return fmt.Errorf("gateway.bind_addr cannot be empty")
	}

	if c.Gateway.ReadBufferSize <= 0 {
		return fmt.Errorf("gateway.read_buffer_size must be positive, got %d", c.Gateway.ReadBufferSize)
	}

	if c.Gateway.WriteBufferSize <= 0 {
		return fmt.Errorf("gateway.write_buffer_size must be positive, got %d", c.Gateway.WriteBufferSize)
	}

	return nil
}

func Default() *Config {
	return &Config{
		Engine: EngineConfig{
			TickRateMs:       25, // 40Hz tick rate
			MaxEntities:      300,
			WorldBounds: Bounds{
				MinX: -1000,
				MinY: -1000,
				MaxX: 1000,
				MaxY: 1000,
			},
			MaxSpeed:        5.0, // units per tick
			AOIRadius:       200.0,
			QuadtreeDepth:   8,
			QuadtreeCapacity: 8,
		},
		Gateway: GatewayConfig{
			BindAddr:          ":8080",
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
			PingPeriod:        54,  // seconds
			PongWait:          60,  // seconds
			WriteWait:         10,  // seconds
			MaxMessageSize:    512, // bytes
			EnableCompression: true,
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
		Postgres: PostgresConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "aether",
			Password:        "password",
			DBName:          "aether",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300, // seconds
		},
	}
}
