package credis

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	redis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/config"
	infras "github.com/gianglt2198/federation-go/package/infras/cache"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/tracing"
)

// Redis represents a Redis client with monitoring
type Redis struct {
	config    config.RedisConfig
	appConfig config.AppConfig

	client *redis.Client
	logger *logging.Logger
}

type RedisParams struct {
	fx.In

	AppConfig config.AppConfig
	Config    config.RedisConfig
	Logger    *logging.Logger
}

type RedisResult struct {
	fx.Out

	Redis *Redis
}

// NewRedis creates a new Redis client
func NewRedis(params RedisParams) RedisResult {
	r := connect(params.AppConfig, params.Config, params.Logger)
	return RedisResult{
		Redis: r,
	}
}

func connect(
	appConfig config.AppConfig,
	redisConfig config.RedisConfig,
	logger *logging.Logger) *Redis {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password:     redisConfig.Password,
		DB:           redisConfig.Database,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	r := &Redis{
		appConfig: appConfig,
		config:    redisConfig,

		client: client,
		logger: logger,
	}

	// Add hooks for monitoring
	client.AddHook(&redisHook{
		logger: logger,
	})

	logger.GetLogger().Info("Redis connection established",
		zap.String("host", redisConfig.Host),
		zap.Int("port", redisConfig.Port),
		zap.Int("database", redisConfig.Database),
		zap.Int("pool_size", redisConfig.PoolSize),
	)

	return r
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	return r.client.Close()
}

// Ping tests the Redis connection
func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Get retrieves a value from Redis
func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	start := time.Now()
	result := r.client.Get(ctx, key)

	// if r.metrics != nil {
	// 	if result.Err() == redis.Nil {
	// 		r.metrics.RecordCacheMiss(r.appConfig.Name, "string", key)
	// 	} else if result.Err() == nil {
	// 		r.metrics.RecordCacheHit(r.appConfig.Name, "string", key)
	// 	}
	// 	r.metrics.RecordCacheOperation(r.appConfig.Name, "string", "get")
	// }

	if result.Err() == redis.Nil {
		return []byte(""), nil // Key not found
	}

	if r.logger != nil && result.Err() != nil {
		r.logger.GetWrappedLogger(ctx).Error("Redis GET failed",
			zap.String("key", key),
			zap.Error(result.Err()),
			zap.Duration("duration", time.Since(start)),
		)
	}

	val, err := result.Result()
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

// Set stores a value in Redis
func (r *Redis) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	start := time.Now()
	result := r.client.Set(ctx, key, value, expiration)

	if r.logger != nil && result.Err() != nil {
		r.logger.GetWrappedLogger(ctx).Error("Redis SET failed",
			zap.String("key", key),
			zap.Error(result.Err()),
			zap.Duration("duration", time.Since(start)),
		)
	}

	return result.Err()
}

// Delete deletes a key from Redis (implements Cache interface)
func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.Del(ctx, key)
}

// Del deletes keys from Redis
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	start := time.Now()
	result := r.client.Del(ctx, keys...)

	if r.logger != nil && result.Err() != nil {
		r.logger.GetWrappedLogger(ctx).Error("Redis DEL failed",
			zap.Strings("keys", keys),
			zap.Error(result.Err()),
			zap.Duration("duration", time.Since(start)),
		)
	}

	return result.Err()
}

// Exists checks if a key exists in Redis (implements Cache interface)
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.ExistsMultiple(ctx, key)
	return count > 0, err
}

// ExistsMultiple checks if keys exist in Redis
func (r *Redis) ExistsMultiple(ctx context.Context, keys ...string) (int64, error) {
	start := time.Now()
	result := r.client.Exists(ctx, keys...)

	if r.logger != nil && result.Err() != nil {
		r.logger.GetWrappedLogger(ctx).Error("Redis EXISTS failed",
			zap.Strings("keys", keys),
			zap.Error(result.Err()),
			zap.Duration("duration", time.Since(start)),
		)
	}

	return result.Result()
}

// HGet retrieves a field from a hash
func (r *Redis) HGet(ctx context.Context, key, field string) (string, error) {
	start := time.Now()
	result := r.client.HGet(ctx, key, field)

	tracing.Meter("redis")

	if result.Err() == redis.Nil {
		return "", nil // Field not found
	}

	if r.logger != nil && result.Err() != nil {
		r.logger.GetWrappedLogger(ctx).Error("Redis HGET failed",
			zap.String("key", key),
			zap.String("field", field),
			zap.Error(result.Err()),
			zap.Duration("duration", time.Since(start)),
		)
	}

	return result.Result()
}

// HSet stores a field in a hash
func (r *Redis) HSet(ctx context.Context, key string, values ...interface{}) error {
	start := time.Now()
	result := r.client.HSet(ctx, key, values...)

	if r.logger != nil && result.Err() != nil {
		r.logger.GetWrappedLogger(ctx).Error("Redis HSET failed",
			zap.String("key", key),
			zap.Error(result.Err()),
			zap.Duration("duration", time.Since(start)),
		)
	}

	return result.Err()
}

// HDel deletes fields from a hash
func (r *Redis) HDel(ctx context.Context, key string, fields ...string) error {
	start := time.Now()
	result := r.client.HDel(ctx, key, fields...)

	if r.logger != nil && result.Err() != nil {
		r.logger.GetWrappedLogger(ctx).Error("Redis HDEL failed",
			zap.String("key", key),
			zap.Strings("fields", fields),
			zap.Error(result.Err()),
			zap.Duration("duration", time.Since(start)),
		)
	}

	return result.Err()
}

// Expire sets a timeout on a key
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	start := time.Now()
	result := r.client.Expire(ctx, key, expiration)

	if r.logger != nil && result.Err() != nil {
		r.logger.GetWrappedLogger(ctx).Error("Redis EXPIRE failed",
			zap.String("key", key),
			zap.Duration("expiration", expiration),
			zap.Error(result.Err()),
			zap.Duration("duration", time.Since(start)),
		)
	}

	return result.Err()
}

// MGet retrieves multiple values from Redis
func (r *Redis) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	result := r.client.MGet(ctx, keys...)
	if result.Err() != nil {
		return nil, result.Err()
	}

	values := make(map[string][]byte)
	for i, val := range result.Val() {
		if val != nil {
			if str, ok := val.(string); ok {
				values[keys[i]] = []byte(str)
			}
		}
	}
	return values, nil
}

// MSet stores multiple values in Redis
func (r *Redis) MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	pipe := r.client.Pipeline()
	for key, value := range items {
		pipe.Set(ctx, key, value, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// MDelete deletes multiple keys from Redis
func (r *Redis) MDelete(ctx context.Context, keys []string) error {
	return r.Del(ctx, keys...)
}

// TTL returns the time to live for a key
func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	result := r.client.TTL(ctx, key)
	return result.Val(), result.Err()
}

// Keys returns keys matching a pattern
func (r *Redis) Keys(ctx context.Context, pattern string) ([]string, error) {
	result := r.client.Keys(ctx, pattern)
	return result.Val(), result.Err()
}

// Scan returns an iterator for keys matching a pattern
func (r *Redis) Scan(ctx context.Context, pattern string) infras.Iterator {
	return &redisIterator{
		client:  r.client,
		pattern: pattern,
		ctx:     ctx,
	}
}

// Lock creates a distributed lock
func (r *Redis) Lock(ctx context.Context, key string, ttl time.Duration) (infras.Lock, error) {
	return &redisLock{
		client: r.client,
		key:    key,
		ttl:    ttl,
	}, nil
}

// HealthCheck returns a health check function for Redis
func (r *Redis) HealthCheck() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return r.Ping(ctx)
	}
}

func (r *Redis) GetClient() *redis.Client {
	return r.client
}

// redisIterator implements infras.Iterator for Redis SCAN
type redisIterator struct {
	client  *redis.Client
	pattern string
	ctx     context.Context
	cursor  uint64
	keys    []string
	index   int
	err     error
}

func (it *redisIterator) Next(ctx context.Context) bool {
	if it.index < len(it.keys) {
		it.index++
		return it.index <= len(it.keys)
	}

	result := it.client.Scan(ctx, it.cursor, it.pattern, 10)
	if result.Err() != nil {
		it.err = result.Err()
		return false
	}

	keys, cursor := result.Val()
	it.keys = append(it.keys, keys...)
	it.cursor = cursor
	it.index++

	return it.index <= len(it.keys) || cursor != 0
}

func (it *redisIterator) Key() string {
	if it.index > 0 && it.index <= len(it.keys) {
		return it.keys[it.index-1]
	}
	return ""
}

func (it *redisIterator) Value() []byte {
	key := it.Key()
	if key == "" {
		return nil
	}

	result := it.client.Get(it.ctx, key)
	if result.Err() != nil {
		return nil
	}
	return []byte(result.Val())
}

func (it *redisIterator) Error() error {
	return it.err
}

// redisLock implements infras.Lock for Redis
type redisLock struct {
	client *redis.Client
	key    string
	ttl    time.Duration
}

func (l *redisLock) Unlock(ctx context.Context) error {
	return l.client.Del(ctx, l.key).Err()
}

func (l *redisLock) Refresh(ctx context.Context, ttl time.Duration) error {
	return l.client.Expire(ctx, l.key, ttl).Err()
}

// redisHook implements redis.Hook for monitoring
type redisHook struct {
	logger *logging.Logger
	metric *redisMetric
}

func (h *redisHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		start := time.Now()
		conn, err := next(ctx, network, addr)

		if h.logger != nil {
			if err != nil {
				h.logger.GetWrappedLogger(ctx).Error("Redis dial failed",
					zap.String("network", network),
					zap.String("addr", addr),
					zap.Error(err),
					zap.Int64("duration", time.Since(start).Milliseconds()),
				)
			} else {
				h.logger.GetWrappedLogger(ctx).Debug("Redis dial successful",
					zap.String("network", network),
					zap.String("addr", addr),
					zap.Int64("duration", time.Since(start).Milliseconds()),
				)
			}
		}

		return conn, err
	}
}

func (h *redisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		duration := time.Since(start).Milliseconds()

		if h.metric != nil {
			h.metric.CacheOperations.Add(ctx, 1)
		}

		if h.logger != nil {
			h.logger.GetWrappedLogger(ctx).Debug("Redis command executed",
				zap.String("command", cmd.Name()),
				zap.String("args", fmt.Sprintf("%v", cmd.Args())),
				zap.Error(err),
				zap.Int64("duration", duration),
			)
		}

		return err
	}
}

func (h *redisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		duration := time.Since(start).Milliseconds()

		if h.metric != nil {
			h.metric.CacheOperations.Add(ctx, 1)
		}

		if h.logger != nil {
			h.logger.GetWrappedLogger(ctx).Debug("Redis pipeline executed",
				zap.Int("commands_count", len(cmds)),
				zap.Error(err),
				zap.Int64("duration", duration),
			)
		}

		return err
	}
}
