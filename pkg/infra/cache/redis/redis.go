package credis

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/gianglt2198/graphql-gateway-go/pkg/config"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/cache"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/monitoring"
	"github.com/google/uuid"
	redis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type redisCache struct {
	client    *redis.Client
	config    config.RedisConfig
	namespace string
}

var _ cache.Cache = (*redisCache)(nil)

type RedisParams struct {
	fx.In

	Log *monitoring.AppLogger
	Cfg config.RedisConfig
}

func New(params RedisParams) *redisCache {
	provider := connect(params.Log, params.Cfg)
	return provider
}

func connect(log *monitoring.AppLogger, cfg config.RedisConfig) *redisCache {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		log.GetLogger().Fatal("Failed to parse Redis URL", zap.Error(err))
	}

	if cfg.TLSEnabled {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	opts.MaxRetries = 60
	opts.MinRetryBackoff = time.Second
	opts.MaxRetryBackoff = time.Second * 120

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	go func() {
		if err := client.Ping(ctx).Err(); err != nil {
			log.GetLogger().Fatal("Redis connection error", zap.Error(err))
		}
	}()
	opts.OnConnect = func(ctx context.Context, cn *redis.Conn) error {
		cancel()
		log.GetLogger().Info("Connected redis successfully")
		return nil
	}

	return &redisCache{
		client:    client,
		config:    cfg,
		namespace: cfg.Namespace,
	}
}

func (c *redisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *redisCache) Close() error {
	return c.client.Close()
}

// Basic Redis implementation
func (c *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, c.prefixKey(key)).Bytes()
	if err == redis.Nil {
		return nil, cache.ErrKeyNotFound
	} else if err != nil {
		return nil, err
	}

	return val, nil
}

func (c *redisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, c.prefixKey(key), value, ttl).Err()
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.prefixKey(key)).Err()
}

func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, c.prefixKey(key)).Result()
	return n > 0, err
}

// Batch operations
func (c *redisCache) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefixKey(key)
	}

	vals, err := c.client.MGet(ctx, prefixedKeys...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte)
	for i, val := range vals {
		if val != nil {
			if bytes, ok := val.([]byte); ok {
				result[keys[i]] = bytes
			}
		}
	}
	return result, nil
}

func (c *redisCache) MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	pipe := c.client.Pipeline()
	for key, value := range items {
		pipe.Set(ctx, c.prefixKey(key), value, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisCache) MDelete(ctx context.Context, keys []string) error {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefixKey(key)
	}
	return c.client.Del(ctx, prefixedKeys...).Err()
}

// TTL operations
func (c *redisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, c.prefixKey(key)).Result()
}

func (c *redisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, c.prefixKey(key), ttl).Err()
}

// Pattern operations
func (c *redisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, c.prefixKey(pattern)).Result()
}

type redisIterator struct {
	redisCache *redisCache
	pattern    string
	cursor     uint64
	keys       []string
	values     map[string][]byte
	current    int
	err        error
}

func (c *redisCache) Scan(ctx context.Context, pattern string) cache.Iterator {
	return &redisIterator{
		redisCache: c,
		pattern:    c.prefixKey(pattern),
		cursor:     0,
		current:    -1,
	}
}

func (it *redisIterator) Next(ctx context.Context) bool {
	if it.current < len(it.keys)-1 {
		it.current++
		return true
	}

	keys, cursor, err := it.redisCache.client.Scan(ctx, it.cursor, it.pattern, int64(it.redisCache.config.ScanCount)).Result()
	if err != nil {
		it.err = err
		return false
	}

	if len(keys) == 0 && cursor == 0 {
		return false
	}

	it.values, err = it.redisCache.MGet(ctx, keys)
	if err != nil {
		it.err = err
		return false
	}

	it.keys = keys
	it.cursor = cursor
	it.current = 0
	return len(it.keys) > 0
}

func (it *redisIterator) Key() string {
	if it.current < 0 || it.current >= len(it.keys) {
		return ""
	}
	return it.keys[it.current]
}

func (it *redisIterator) Value() []byte {
	if it.current < 0 || it.current >= len(it.values) {
		return nil
	}
	if v, ok := it.values[it.keys[it.current]]; ok {
		return v
	}
	return nil
}

func (it *redisIterator) Error() error {
	return it.err
}

// Lock implementation
type redisLock struct {
	cache *redisCache
	key   string
	token string
}

func (c *redisCache) Lock(ctx context.Context, key string, ttl time.Duration) (cache.Lock, error) {
	lockKey := c.prefixKey("lock:" + key)
	token := uuid.New().String()

	ok, err := c.client.SetNX(ctx, lockKey, token, ttl).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, cache.ErrLockAcquired
	}

	return &redisLock{
		cache: c,
		key:   lockKey,
		token: token,
	}, nil
}

func (l *redisLock) Unlock(ctx context.Context) error {
	script := `  
        if redis.call("get", KEYS[1]) == ARGV[1] then  
            return redis.call("del", KEYS[1])  
        else  
            return 0  
        end`

	result := l.cache.client.Eval(ctx, script, []string{l.key}, l.token)
	return result.Err()
}

func (l *redisLock) Refresh(ctx context.Context, ttl time.Duration) error {
	script := `  
        if redis.call("get", KEYS[1]) == ARGV[1] then  
            return redis.call("pexpire", KEYS[1], ARGV[2])  
        else  
            return 0  
        end`

	result := l.cache.client.Eval(ctx, script, []string{l.key}, l.token, ttl.Milliseconds())
	if result.Val() == 0 {
		return cache.ErrLockExpired
	}
	return result.Err()
}
