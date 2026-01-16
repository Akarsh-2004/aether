package observability

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Metrics struct {
	logger *zap.Logger

	// Engine metrics
	TickDuration prometheus.Histogram
	EntityCount  prometheus.Gauge
	AOIQueries   prometheus.Counter
	BroadcastSize prometheus.Histogram

	// Gateway metrics
	ActiveConnections prometheus.Gauge
	MessagesReceived  prometheus.Counter
	MessagesSent      prometheus.Counter
	ConnectionErrors   prometheus.Counter

	// Persistence metrics
	RedisOperations    prometheus.CounterVec
	PostgresOperations prometheus.CounterVec
	OutboxEvents       prometheus.Counter

	// System metrics
	MemoryUsage prometheus.Gauge
	GCCount     prometheus.Counter
}

func NewMetrics(logger *zap.Logger) *Metrics {
	return &Metrics{
		logger: logger,
		TickDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "aether_tick_duration_seconds",
			Help:    "Duration of engine tick processing",
			Buckets: prometheus.DefBuckets,
		}),
		EntityCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "aether_entity_count",
			Help: "Number of active entities",
		}),
		AOIQueries: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aether_aoi_queries_total",
			Help: "Total number of AOI queries",
		}),
		BroadcastSize: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "aether_broadcast_size_bytes",
			Help:    "Size of broadcast messages in bytes",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000},
		}),
		ActiveConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "aether_active_connections",
			Help: "Number of active WebSocket connections",
		}),
		MessagesReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aether_messages_received_total",
			Help: "Total number of messages received from clients",
		}),
		MessagesSent: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aether_messages_sent_total",
			Help: "Total number of messages sent to clients",
		}),
		ConnectionErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aether_connection_errors_total",
			Help: "Total number of connection errors",
		}),
		RedisOperations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "aether_redis_operations_total",
			Help: "Total number of Redis operations",
		}, []string{"operation", "status"}),
		PostgresOperations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "aether_postgres_operations_total",
			Help: "Total number of PostgreSQL operations",
		}, []string{"operation", "status"}),
		OutboxEvents: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aether_outbox_events_total",
			Help: "Total number of outbox events processed",
		}),
		MemoryUsage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "aether_memory_usage_bytes",
			Help: "Current memory usage in bytes",
		}),
		GCCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aether_gc_count_total",
			Help: "Total number of garbage collections",
		}),
	}
}

func (m *Metrics) Register() error {
	// Register all metrics with the default registry
	metrics := []prometheus.Collector{
		m.TickDuration,
		m.EntityCount,
		m.AOIQueries,
		m.BroadcastSize,
		m.ActiveConnections,
		m.MessagesReceived,
		m.MessagesSent,
		m.ConnectionErrors,
		m.RedisOperations,
		m.PostgresOperations,
		m.OutboxEvents,
		m.MemoryUsage,
		m.GCCount,
	}

	for _, metric := range metrics {
		if err := prometheus.DefaultRegisterer.Register(metric); err != nil {
			m.logger.Error("Failed to register metric", zap.Error(err))
			return err
		}
	}

	m.logger.Info("Metrics registered successfully")
	return nil
}

func (m *Metrics) StartMetricsServer(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	m.logger.Info("Starting metrics server", zap.String("addr", addr))
	return server.ListenAndServe()
}

// Metric recording helpers

func (m *Metrics) RecordTickDuration(duration time.Duration) {
	m.TickDuration.Observe(duration.Seconds())
}

func (m *Metrics) SetEntityCount(count int) {
	m.EntityCount.Set(float64(count))
}

func (m *Metrics) IncrementAOIQueries() {
	m.AOIQueries.Inc()
}

func (m *Metrics) RecordBroadcastSize(size int) {
	m.BroadcastSize.Observe(float64(size))
}

func (m *Metrics) SetActiveConnections(count int) {
	m.ActiveConnections.Set(float64(count))
}

func (m *Metrics) IncrementMessagesReceived() {
	m.MessagesReceived.Inc()
}

func (m *Metrics) IncrementMessagesSent() {
	m.MessagesSent.Inc()
}

func (m *Metrics) IncrementConnectionErrors() {
	m.ConnectionErrors.Inc()
}

func (m *Metrics) RecordRedisOperation(operation, status string) {
	m.RedisOperations.WithLabelValues(operation, status).Inc()
}

func (m *Metrics) RecordPostgresOperation(operation, status string) {
	m.PostgresOperations.WithLabelValues(operation, status).Inc()
}

func (m *Metrics) IncrementOutboxEvents() {
	m.OutboxEvents.Inc()
}

func (m *Metrics) SetMemoryUsage(bytes int64) {
	m.MemoryUsage.Set(float64(bytes))
}

func (m *Metrics) IncrementGCCount() {
	m.GCCount.Inc()
}
