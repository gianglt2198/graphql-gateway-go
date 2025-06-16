package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gianglt2198/federation-go/package/config"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPRequestsInFlight  *prometheus.GaugeVec

	// GraphQL metrics
	GraphQLRequestsTotal    *prometheus.CounterVec
	GraphQLRequestDuration  *prometheus.HistogramVec
	GraphQLRequestsInFlight *prometheus.GaugeVec
	GraphQLOperationsTotal  *prometheus.CounterVec
	GraphQLErrorsTotal      *prometheus.CounterVec

	// Database metrics
	DatabaseConnectionsActive *prometheus.GaugeVec
	DatabaseConnectionsIdle   *prometheus.GaugeVec
	DatabaseQueriesTotal      *prometheus.CounterVec
	DatabaseQueryDuration     *prometheus.HistogramVec

	// Federation metrics
	SubgraphRequestsTotal    *prometheus.CounterVec
	SubgraphRequestDuration  *prometheus.HistogramVec
	SubgraphErrorsTotal      *prometheus.CounterVec
	SchemaCompositionTotal   *prometheus.CounterVec
	SchemaCompositionErrors  *prometheus.CounterVec

	// Service metrics
	ServiceHealthStatus *prometheus.GaugeVec
	ServiceUptime       *prometheus.GaugeVec

	// Cache metrics
	CacheHitsTotal   *prometheus.CounterVec
	CacheMissesTotal *prometheus.CounterVec
	CacheOperations  *prometheus.CounterVec

	// Message queue metrics
	MessagesSentTotal     *prometheus.CounterVec
	MessagesReceivedTotal *prometheus.CounterVec
	MessageProcessingTime *prometheus.HistogramVec

	registry *prometheus.Registry
}

// NewMetrics creates a new metrics instance
func NewMetrics(config config.MetricsConfig) *Metrics {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"service", "method", "path", "status_code"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"service", "method", "path"},
		),
		HTTPRequestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_requests_in_flight",
				Help:      "Number of HTTP requests currently being processed",
			},
			[]string{"service"},
		),

		// GraphQL metrics
		GraphQLRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "graphql_requests_total",
				Help:      "Total number of GraphQL requests",
			},
			[]string{"service", "operation_type", "operation_name"},
		),
		GraphQLRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "graphql_request_duration_seconds",
				Help:      "GraphQL request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"service", "operation_type", "operation_name"},
		),
		GraphQLRequestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "graphql_requests_in_flight",
				Help:      "Number of GraphQL requests currently being processed",
			},
			[]string{"service"},
		),
		GraphQLOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "graphql_operations_total",
				Help:      "Total number of GraphQL operations executed",
			},
			[]string{"service", "operation_type", "operation_name", "status"},
		),
		GraphQLErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "graphql_errors_total",
				Help:      "Total number of GraphQL errors",
			},
			[]string{"service", "error_type", "operation_name"},
		),

		// Database metrics
		DatabaseConnectionsActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "database_connections_active",
				Help:      "Number of active database connections",
			},
			[]string{"service", "database"},
		),
		DatabaseConnectionsIdle: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "database_connections_idle",
				Help:      "Number of idle database connections",
			},
			[]string{"service", "database"},
		),
		DatabaseQueriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "database_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"service", "database", "operation", "table"},
		),
		DatabaseQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "database_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"service", "database", "operation", "table"},
		),

		// Federation metrics
		SubgraphRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "subgraph_requests_total",
				Help:      "Total number of subgraph requests",
			},
			[]string{"service", "subgraph", "operation"},
		),
		SubgraphRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "subgraph_request_duration_seconds",
				Help:      "Subgraph request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"service", "subgraph", "operation"},
		),
		SubgraphErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "subgraph_errors_total",
				Help:      "Total number of subgraph errors",
			},
			[]string{"service", "subgraph", "error_type"},
		),
		SchemaCompositionTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "schema_composition_total",
				Help:      "Total number of schema compositions",
			},
			[]string{"service", "status"},
		),
		SchemaCompositionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "schema_composition_errors_total",
				Help:      "Total number of schema composition errors",
			},
			[]string{"service", "error_type"},
		),

		// Service metrics
		ServiceHealthStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "service_health_status",
				Help:      "Service health status (1 = healthy, 0 = unhealthy)",
			},
			[]string{"service", "component"},
		),
		ServiceUptime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "service_uptime_seconds",
				Help:      "Service uptime in seconds",
			},
			[]string{"service"},
		),

		// Cache metrics
		CacheHitsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
			[]string{"service", "cache_type", "key_pattern"},
		),
		CacheMissesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
			[]string{"service", "cache_type", "key_pattern"},
		),
		CacheOperations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "cache_operations_total",
				Help:      "Total number of cache operations",
			},
			[]string{"service", "cache_type", "operation"},
		),

		// Message queue metrics
		MessagesSentTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "messages_sent_total",
				Help:      "Total number of messages sent",
			},
			[]string{"service", "subject", "message_type"},
		),
		MessagesReceivedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "messages_received_total",
				Help:      "Total number of messages received",
			},
			[]string{"service", "subject", "message_type"},
		),
		MessageProcessingTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "message_processing_duration_seconds",
				Help:      "Message processing duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"service", "subject", "message_type"},
		),

		registry: registry,
	}

	// Register all metrics
	m.registerMetrics()

	return m
}

// registerMetrics registers all metrics with the registry
func (m *Metrics) registerMetrics() {
	collectors := []prometheus.Collector{
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.HTTPRequestsInFlight,
		m.GraphQLRequestsTotal,
		m.GraphQLRequestDuration,
		m.GraphQLRequestsInFlight,
		m.GraphQLOperationsTotal,
		m.GraphQLErrorsTotal,
		m.DatabaseConnectionsActive,
		m.DatabaseConnectionsIdle,
		m.DatabaseQueriesTotal,
		m.DatabaseQueryDuration,
		m.SubgraphRequestsTotal,
		m.SubgraphRequestDuration,
		m.SubgraphErrorsTotal,
		m.SchemaCompositionTotal,
		m.SchemaCompositionErrors,
		m.ServiceHealthStatus,
		m.ServiceUptime,
		m.CacheHitsTotal,
		m.CacheMissesTotal,
		m.CacheOperations,
		m.MessagesSentTotal,
		m.MessagesReceivedTotal,
		m.MessageProcessingTime,
	}

	for _, collector := range collectors {
		m.registry.MustRegister(collector)
	}
}

// GetRegistry returns the Prometheus registry
func (m *Metrics) GetRegistry() *prometheus.Registry {
	return m.registry
}

// Handler returns the HTTP handler for metrics endpoint
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// RecordHTTPRequest records HTTP request metrics
func (m *Metrics) RecordHTTPRequest(service, method, path string, statusCode int, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(service, method, path, strconv.Itoa(statusCode)).Inc()
	m.HTTPRequestDuration.WithLabelValues(service, method, path).Observe(duration.Seconds())
}

// RecordGraphQLRequest records GraphQL request metrics
func (m *Metrics) RecordGraphQLRequest(service, operationType, operationName string, duration time.Duration, hasErrors bool) {
	m.GraphQLRequestsTotal.WithLabelValues(service, operationType, operationName).Inc()
	m.GraphQLRequestDuration.WithLabelValues(service, operationType, operationName).Observe(duration.Seconds())
	
	status := "success"
	if hasErrors {
		status = "error"
	}
	m.GraphQLOperationsTotal.WithLabelValues(service, operationType, operationName, status).Inc()
}

// RecordGraphQLError records GraphQL error metrics
func (m *Metrics) RecordGraphQLError(service, errorType, operationName string) {
	m.GraphQLErrorsTotal.WithLabelValues(service, errorType, operationName).Inc()
}

// RecordDatabaseQuery records database query metrics
func (m *Metrics) RecordDatabaseQuery(service, database, operation, table string, duration time.Duration) {
	m.DatabaseQueriesTotal.WithLabelValues(service, database, operation, table).Inc()
	m.DatabaseQueryDuration.WithLabelValues(service, database, operation, table).Observe(duration.Seconds())
}

// RecordSubgraphRequest records subgraph request metrics
func (m *Metrics) RecordSubgraphRequest(service, subgraph, operation string, duration time.Duration) {
	m.SubgraphRequestsTotal.WithLabelValues(service, subgraph, operation).Inc()
	m.SubgraphRequestDuration.WithLabelValues(service, subgraph, operation).Observe(duration.Seconds())
}

// RecordSubgraphError records subgraph error metrics
func (m *Metrics) RecordSubgraphError(service, subgraph, errorType string) {
	m.SubgraphErrorsTotal.WithLabelValues(service, subgraph, errorType).Inc()
}

// RecordSchemaComposition records schema composition metrics
func (m *Metrics) RecordSchemaComposition(service string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	m.SchemaCompositionTotal.WithLabelValues(service, status).Inc()
}

// RecordSchemaCompositionError records schema composition error metrics
func (m *Metrics) RecordSchemaCompositionError(service, errorType string) {
	m.SchemaCompositionErrors.WithLabelValues(service, errorType).Inc()
}

// SetServiceHealth sets service health status
func (m *Metrics) SetServiceHealth(service, component string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.ServiceHealthStatus.WithLabelValues(service, component).Set(value)
}

// SetServiceUptime sets service uptime
func (m *Metrics) SetServiceUptime(service string, uptime time.Duration) {
	m.ServiceUptime.WithLabelValues(service).Set(uptime.Seconds())
}

// RecordCacheHit records cache hit metrics
func (m *Metrics) RecordCacheHit(service, cacheType, keyPattern string) {
	m.CacheHitsTotal.WithLabelValues(service, cacheType, keyPattern).Inc()
}

// RecordCacheMiss records cache miss metrics
func (m *Metrics) RecordCacheMiss(service, cacheType, keyPattern string) {
	m.CacheMissesTotal.WithLabelValues(service, cacheType, keyPattern).Inc()
}

// RecordCacheOperation records cache operation metrics
func (m *Metrics) RecordCacheOperation(service, cacheType, operation string) {
	m.CacheOperations.WithLabelValues(service, cacheType, operation).Inc()
}

// RecordMessageSent records message sent metrics
func (m *Metrics) RecordMessageSent(service, subject, messageType string) {
	m.MessagesSentTotal.WithLabelValues(service, subject, messageType).Inc()
}

// RecordMessageReceived records message received metrics
func (m *Metrics) RecordMessageReceived(service, subject, messageType string) {
	m.MessagesReceivedTotal.WithLabelValues(service, subject, messageType).Inc()
}

// RecordMessageProcessing records message processing metrics
func (m *Metrics) RecordMessageProcessing(service, subject, messageType string, duration time.Duration) {
	m.MessageProcessingTime.WithLabelValues(service, subject, messageType).Observe(duration.Seconds())
}

// IncHTTPRequestsInFlight increments HTTP requests in flight
func (m *Metrics) IncHTTPRequestsInFlight(service string) {
	m.HTTPRequestsInFlight.WithLabelValues(service).Inc()
}

// DecHTTPRequestsInFlight decrements HTTP requests in flight
func (m *Metrics) DecHTTPRequestsInFlight(service string) {
	m.HTTPRequestsInFlight.WithLabelValues(service).Dec()
}

// IncGraphQLRequestsInFlight increments GraphQL requests in flight
func (m *Metrics) IncGraphQLRequestsInFlight(service string) {
	m.GraphQLRequestsInFlight.WithLabelValues(service).Inc()
}

// DecGraphQLRequestsInFlight decrements GraphQL requests in flight
func (m *Metrics) DecGraphQLRequestsInFlight(service string) {
	m.GraphQLRequestsInFlight.WithLabelValues(service).Dec()
}

// SetDatabaseConnections sets database connection metrics
func (m *Metrics) SetDatabaseConnections(service, database string, active, idle int) {
	m.DatabaseConnectionsActive.WithLabelValues(service, database).Set(float64(active))
	m.DatabaseConnectionsIdle.WithLabelValues(service, database).Set(float64(idle))
} 