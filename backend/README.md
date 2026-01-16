# Aether - High-Performance Realtime Spatial Engine

A production-grade, authoritative multiplayer backend built for low-latency spatial interactions. Designed for 100-300 concurrent entities with sub-50ms end-to-end latency.

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client        │    │   Client        │    │   Client        │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │   WebSocket Gateway        │
                    │   (Binary + Protobuf)      │
                    └─────────────┬─────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │   Spatial Engine           │
                    │   • Fixed Tick Loop        │
                    │   • Quadtree Index         │
                    │   • AOI Management        │
                    │   • Authority System      │
                    └─────────────┬─────────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                      │
┌─────────▼─────────┐  ┌─────────▼─────────┐  ┌─────────▼─────────┐
│      Redis        │  │    PostgreSQL     │  │   Prometheus     │
│  • Presence       │  │  • Snapshots      │  │   • Metrics      │
│  • Heartbeats     │  │  • Outbox         │  │   • Monitoring  │
│  • Sessions       │  │  • Sessions       │  │                  │
└───────────────────┘  └───────────────────┘  └───────────────────┘
```

## Core Principles

- **Authoritative Server**: Server owns all game state, clients send intents only
- **Deterministic Processing**: Fixed 25ms tick loop (40Hz) for predictable behavior
- **Binary Protocol**: Protobuf over WebSockets for minimal overhead
- **Spatial Optimization**: Quadtree indexing with AOI (Area of Interest) culling
- **No Global Locks**: Concurrent-safe design with minimal contention
- **Graceful Degradation**: Backpressure handling and packet dropping under load

## Performance Targets

- **Concurrent Entities**: 100-300
- **End-to-End Latency**: <50ms
- **Tick Rate**: 40Hz (25ms fixed timestep)
- **Message Size**: <512 bytes
- **Memory Usage**: <512MB for 300 entities

## Quick Start

### Prerequisites

- Go 1.21+
- Redis 6.0+
- PostgreSQL 13+
- Docker & Docker Compose (optional)

### Local Development

1. **Clone and build**:
```bash
git clone <repository>
cd aether/backend
go mod tidy
go build -o aether-server ./cmd/server
```

2. **Start dependencies**:
```bash
# Using Docker Compose (recommended)
cd docker
docker-compose up -d redis postgres

# Or start manually
redis-server
# Start PostgreSQL with aether database
```

3. **Run the server**:
```bash
./aether-server -config config.yaml
```

### Docker Deployment

```bash
cd docker
docker-compose up -d
```

Services will be available at:
- Aether Server: `http://localhost:8080`
- Redis: `localhost:6379`
- PostgreSQL: `localhost:5432`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (admin/admin)

## Configuration

Key configuration options in `config.yaml`:

```yaml
engine:
  tick_rate_ms: 25          # Fixed timestep (40Hz)
  max_entities: 300         # Max concurrent entities
  world_bounds:             # World boundaries
    min_x: -1000
    min_y: -1000
    max_x: 1000
    max_y: 1000
  max_speed: 5.0            # Max movement per tick
  aoi_radius: 200.0         # Area of Interest radius

gateway:
  bind_addr: ":8080"        # WebSocket bind address
  max_message_size: 512     # Max message size
  enable_compression: true  # WebSocket compression
```

## Protocol

### Message Flow

1. **Client → Server**: Movement intents, spawn requests, heartbeats
2. **Server → Client**: Entity states, corrections, AOI updates

### Key Messages

```protobuf
// Client movement intent
message MovementDelta {
  uint32 entity_id = 1;
  uint64 sequence = 2;        // Client-side sequence
  float delta_x = 3;          // Movement delta
  float delta_y = 4;
  uint64 timestamp = 5;
}

// Server correction when client prediction fails
message Correction {
  uint32 entity_id = 1;
  float correct_x = 2;
  float correct_y = 3;
  uint64 ack_sequence = 6;    // Last acknowledged sequence
}
```

## Architecture Components

### Spatial Engine

**Fixed Tick Loop**: Deterministic 25ms timestep ensures consistent physics and movement validation.

**Quadtree Index**: Efficient spatial queries with O(log n) complexity for entity lookup and AOI calculations.

**AOI Management**: Only sends relevant entity updates to clients within their Area of Interest, minimizing bandwidth.

**Authority System**: Server validates all movement, prevents speed hacks and teleportation, generates corrections.

### WebSocket Gateway

**Binary Protocol**: Uses Protobuf for compact, fast serialization.

**Backpressure Handling**: Non-blocking sends with packet dropping for slow clients.

**Connection Management**: Ping/pong health checks, graceful shutdown, connection pooling.

### Persistence Layer

**Redis**: Ephemeral data - presence, heartbeats, active sessions with TTL.

**PostgreSQL**: Persistent data - entity snapshots, outbox events, session history.

**Outbox Pattern**: Reliable event delivery with transactional guarantees.

### Observability

**Prometheus Metrics**: Tick duration, entity counts, AOI queries, message rates.

**Structured Logging**: JSON logs with correlation IDs for debugging.

**Health Checks**: Database connectivity, memory usage, tick processing health.

## Development Guidelines

### Performance Rules

- **No JSON on hot path**: Use Protobuf for all client communication
- **Minimize allocations**: Reuse buffers, object pools for frequent operations
- **Avoid global locks**: Use sync.Map, atomic operations, or sharding
- **Batch operations**: Group database writes, use connection pooling

### Code Organization

```
internal/
├── engine/          # Core spatial engine
│   ├── tick/       # Fixed timestep loop
│   ├── spatial/    # Quadtree implementation
│   ├── entity/     # Entity management
│   └── aoi/        # Area of Interest
├── gateway/        # WebSocket handling
├── protocol/       # Protobuf codec
├── persistence/    # Database layer
└── observability/  # Metrics & logging
```

### Testing

```bash
# Run unit tests
go test ./...

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...

# Integration tests (requires dependencies)
go test -tags=integration ./...
```

## Scaling Considerations

### Vertical Scaling

- **CPU**: Tick processing is CPU-bound, prioritize single-thread performance
- **Memory**: ~1.5KB per entity, 300 entities ≈ 450KB base usage
- **Network**: ~100KB/s per 100 active entities at 40Hz

### Horizontal Scaling

Future scaling strategies:

1. **Spatial Sharding**: Partition world across multiple server instances
2. **Entity Migration**: Handoff entities between shard boundaries
3. **Load Balancing**: Distribute connections based on geographic location
4. **Multi-Region**: Deploy edge servers for reduced latency

### Performance Optimizations

- **SIMD Instructions**: Vectorized position calculations
- **Memory Layout**: Structure of arrays for better cache locality
- **Batch Processing**: Group entity updates for database writes
- **Compression**: Enable WebSocket compression for high-latency clients

## Monitoring

### Key Metrics

- `aether_tick_duration_seconds`: Tick processing time
- `aether_entity_count`: Active entities
- `aether_aoi_queries_total`: AOI query frequency
- `aether_active_connections`: WebSocket connections
- `aether_messages_sent_total`: Message throughput

### Alerting

Recommended alerts:
- Tick duration > 15ms (60% of tick budget)
- Entity count > 270 (90% capacity)
- Connection errors > 10/min
- Memory usage > 400MB

## Troubleshooting

### Common Issues

**High Tick Latency**:
- Check quadtree depth and entity distribution
- Monitor AOI query complexity
- Profile CPU usage during peak load

**Connection Drops**:
- Verify WebSocket buffer sizes
- Check client read/write timeouts
- Monitor network latency

**Memory Leaks**:
- Profile entity lifecycle management
- Check goroutine leaks in connection handlers
- Monitor buffer pool usage

### Debug Tools

```bash
# Enable debug logging
export LOG_LEVEL=debug

# Profile CPU usage
go tool pprof http://localhost:8080/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:8080/debug/pprof/heap

# Goroutine inspection
curl http://localhost:8080/debug/pprof/goroutine?debug=1
```

## Contributing

1. **Follow Go conventions**: Use `gofmt`, proper error handling
2. **Write tests**: Unit tests for all core components
3. **Benchmark changes**: Ensure no performance regressions
4. **Document decisions**: Explain WHY, not just WHAT
5. **Update README**: Keep documentation current

## License

[License information]

## Contact

[Contact information]
