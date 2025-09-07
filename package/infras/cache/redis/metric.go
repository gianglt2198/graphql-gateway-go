package redis

import (
	"log"

	"go.opentelemetry.io/otel/metric"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/tracing"
)

type redisMetric struct {
	config    config.RedisConfig
	appConfig config.AppConfig

	// Cache metrics
	CacheHitsTotal   metric.Int64Counter
	CacheMissesTotal metric.Int64Counter
	CacheOperations  metric.Int64Counter
}

func NewMetrics(config config.RedisConfig, appConfig config.AppConfig) *redisMetric {
	rm := &redisMetric{
		config:    config,
		appConfig: appConfig,
	}

	m := tracing.Meter(appConfig.Name)

	var err error

	rm.CacheHitsTotal, err = m.Int64Counter(
		"cache_hit_total",
		metric.WithDescription("Total number of hit cach key received."),
		metric.WithUnit("{requests}"),
	)
	if err != nil {
		log.Fatalf("creating meter cache hit counter failed: %v", err)
	}
	rm.CacheMissesTotal, err = m.Int64Counter(
		"cache_miss_total",
		metric.WithDescription("Total number of miss cache key received."),
		metric.WithUnit("{requests}"),
	)
	if err != nil {
		log.Fatalf("creating meter cache miss counter failed: %v", err)
	}
	rm.CacheOperations, err = m.Int64Counter(
		"cache_operation_total",
		metric.WithDescription("Total number of cache operation."),
		metric.WithUnit("{requests}"),
	)
	if err != nil {
		log.Fatalf("creating meter cache operation counter failed: %v", err)
	}

	return rm
}
