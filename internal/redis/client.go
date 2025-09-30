package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/logger"
	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client with our configuration
type Client struct {
	rdb    *redis.Client
	config *config.RedisConfig
}

// NewClient creates a new Redis client
func NewClient(cfg *config.RedisConfig) (*Client, error) {
	if !cfg.Enabled {
		logger.LogWarn(logger.ServiceRedis, "Redis is disabled")
		return nil, nil
	}

	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Override with config values
	if cfg.Password != "" {
		opts.Password = cfg.Password
	}
	opts.DB = cfg.DB
	opts.MaxRetries = cfg.MaxRetries
	opts.DialTimeout = cfg.DialTimeout
	opts.ReadTimeout = cfg.ReadTimeout
	opts.WriteTimeout = cfg.WriteTimeout
	opts.PoolSize = cfg.PoolSize
	opts.MinIdleConns = cfg.MinIdleConns

	rdb := redis.NewClient(opts)

	client := &Client{
		rdb:    rdb,
		config: cfg,
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.LogInfo(logger.ServiceRedis, "Redis client connected successfully", map[string]interface{}{
		"url": cfg.URL,
		"db":  cfg.DB,
	})

	return client, nil
}

// Ping tests the Redis connection
func (c *Client) Ping(ctx context.Context) error {
	if c == nil {
		return fmt.Errorf("Redis client is disabled")
	}
	return c.rdb.Ping(ctx).Err()
}

// Close closes the Redis connection
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	return c.rdb.Close()
}

// GetClient returns the underlying Redis client
func (c *Client) GetClient() *redis.Client {
	return c.rdb
}

// Publish publishes a message to a channel
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	if c == nil {
		return fmt.Errorf("Redis client is disabled")
	}

	start := time.Now()
	err := c.rdb.Publish(ctx, channel, message).Err()
	duration := time.Since(start)

	if err != nil {
		logger.LogError(logger.ServiceRedis, "Failed to publish message", err, map[string]interface{}{
			"channel":  channel,
			"duration": duration.String(),
		})
		return err
	}

	logger.LogDebug(logger.ServiceRedis, "Message published", map[string]interface{}{
		"channel":  channel,
		"duration": duration.String(),
	})

	return nil
}

// Subscribe subscribes to a channel
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if c == nil {
		return nil
	}
	return c.rdb.Subscribe(ctx, channels...)
}

// Set sets a key-value pair with expiration
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if c == nil {
		return fmt.Errorf("Redis client is disabled")
	}

	start := time.Now()
	err := c.rdb.Set(ctx, key, value, expiration).Err()
	duration := time.Since(start)

	if err != nil {
		logger.LogError(logger.ServiceRedis, "Failed to set key", err, map[string]interface{}{
			"key":      key,
			"duration": duration.String(),
		})
		return err
	}

	logger.LogDebug(logger.ServiceRedis, "Key set", map[string]interface{}{
		"key":        key,
		"expiration": expiration.String(),
		"duration":   duration.String(),
	})

	return nil
}

// Get gets a value by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("Redis client is disabled")
	}

	start := time.Now()
	result := c.rdb.Get(ctx, key)
	duration := time.Since(start)

	if err := result.Err(); err != nil {
		if err == redis.Nil {
			logger.LogDebug(logger.ServiceRedis, "Key not found", map[string]interface{}{
				"key":      key,
				"duration": duration.String(),
			})
			return "", nil
		}

		logger.LogError(logger.ServiceRedis, "Failed to get key", err, map[string]interface{}{
			"key":      key,
			"duration": duration.String(),
		})
		return "", err
	}

	logger.LogDebug(logger.ServiceRedis, "Key retrieved", map[string]interface{}{
		"key":      key,
		"duration": duration.String(),
	})

	return result.Val(), nil
}

// Del deletes a key
func (c *Client) Del(ctx context.Context, keys ...string) error {
	if c == nil {
		return fmt.Errorf("Redis client is disabled")
	}

	start := time.Now()
	err := c.rdb.Del(ctx, keys...).Err()
	duration := time.Since(start)

	if err != nil {
		logger.LogError(logger.ServiceRedis, "Failed to delete keys", err, map[string]interface{}{
			"keys":     keys,
			"duration": duration.String(),
		})
		return err
	}

	logger.LogDebug(logger.ServiceRedis, "Keys deleted", map[string]interface{}{
		"keys":     keys,
		"duration": duration.String(),
	})

	return nil
}

// SAdd adds members to a set
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if c == nil {
		return fmt.Errorf("Redis client is disabled")
	}

	start := time.Now()
	err := c.rdb.SAdd(ctx, key, members...).Err()
	duration := time.Since(start)

	if err != nil {
		logger.LogError(logger.ServiceRedis, "Failed to add set members", err, map[string]interface{}{
			"key":      key,
			"duration": duration.String(),
		})
		return err
	}

	logger.LogDebug(logger.ServiceRedis, "Set members added", map[string]interface{}{
		"key":      key,
		"count":    len(members),
		"duration": duration.String(),
	})

	return nil
}

// SMembers gets all members of a set
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	if c == nil {
		return nil, fmt.Errorf("Redis client is disabled")
	}

	start := time.Now()
	result := c.rdb.SMembers(ctx, key)
	duration := time.Since(start)

	if err := result.Err(); err != nil {
		logger.LogError(logger.ServiceRedis, "Failed to get set members", err, map[string]interface{}{
			"key":      key,
			"duration": duration.String(),
		})
		return nil, err
	}

	logger.LogDebug(logger.ServiceRedis, "Set members retrieved", map[string]interface{}{
		"key":      key,
		"count":    len(result.Val()),
		"duration": duration.String(),
	})

	return result.Val(), nil
}

// Expire sets expiration on a key
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if c == nil {
		return fmt.Errorf("Redis client is disabled")
	}

	start := time.Now()
	err := c.rdb.Expire(ctx, key, expiration).Err()
	duration := time.Since(start)

	if err != nil {
		logger.LogError(logger.ServiceRedis, "Failed to set key expiration", err, map[string]interface{}{
			"key":      key,
			"duration": duration.String(),
		})
		return err
	}

	logger.LogDebug(logger.ServiceRedis, "Key expiration set", map[string]interface{}{
		"key":        key,
		"expiration": expiration.String(),
		"duration":   duration.String(),
	})

	return nil
}
