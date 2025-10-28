package redis

import (
	"context"
	"errors"
	"fmt"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/logger"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConnection struct {
	Client redis.UniversalClient
	Prefix string
}

func NewRedisConnections(cfg *config.Config) (*RedisConnection, error) {
	if cfg.REDIS == nil {
		fmt.Println(ErrEmptyConfig)
		logger.Log.Fatal(ErrEmptyConfig)

		return nil, ErrEmptyConfig
	}

	redisDb, err := strconv.Atoi(cfg.REDIS.REDIS_DB)
	if err != nil {
		msg := "invalid redis db value [%s]: %v"
		fmt.Println(fmt.Errorf(msg, cfg.REDIS.REDIS_NAME, err))
		logger.Log.Errorf(msg, cfg.REDIS.REDIS_NAME, err)

		return nil, ErrInvalidDB
	}

	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{cfg.REDIS.REDIS_ADDR},
		Password: cfg.REDIS.REDIS_PASSWORD,
		DB:       redisDb,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		msg := "failed to connect to Redis [%s]: %v"
		fmt.Println(fmt.Errorf(msg, cfg.REDIS.REDIS_NAME, err))
		logger.Log.Errorf(msg, cfg.REDIS.REDIS_NAME, err)

		return nil, ErrConnectionFailed
	}

	return &RedisConnection{
		Client: client,
		Prefix: cfg.REDIS.REDIS_PREFIX,
	}, nil
}

func (rc *RedisConnection) WrapKey(key string) string {
	if strings.HasPrefix(key, rc.Prefix) {
		return key
	}

	return fmt.Sprintf("%s%s", rc.Prefix, key)
}

func (rc *RedisConnection) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	result, err := rc.Client.SetNX(ctx, rc.WrapKey(key), value, expiration).Result()
	if err != nil {
		logger.Log.Errorf("failed to set nx cache: %v", err)
		return false, ErrSetNX
	}

	return result, nil
}

func (rc *RedisConnection) MSetNX(ctx context.Context, keyValues []interface{}) (bool, error) {
	result, err := rc.Client.MSetNX(ctx, keyValues).Result()
	if err != nil {
		logger.Log.Errorf("failed to set multiple nx cache: %v", err)
		return false, ErrSetMNX
	}

	return result, nil
}

func (rc *RedisConnection) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := rc.Client.Set(ctx, rc.WrapKey(key), value, expiration).Err()
	if err != nil {
		logger.Log.Errorf("failed to set cache: %v", err)
		return ErrSetCache
	}

	return nil
}

func (rc *RedisConnection) GetCache(ctx context.Context, key string) (string, error) {
	val, err := rc.Client.Get(ctx, rc.WrapKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}

		logger.Log.Errorf("failed to get cache: %v", err)
		return "", ErrGetCache
	}

	return val, nil
}

func (rc *RedisConnection) DelCache(ctx context.Context, keys ...string) error {
	if len(keys) <= 0 {
		return ErrDelCacheEmptyKeys
	}
	for idx, key := range keys {
		keys[idx] = rc.WrapKey(key)
	}

	if err := rc.Client.Del(ctx, keys...).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}

		logger.Log.Errorf("failed to remove cache [%s]: %v", strings.Join(keys, ", "), err)
		return ErrDelCache
	}

	return nil
}

func (rc *RedisConnection) CacheRemember(ctx context.Context, key string, ttl time.Duration, fetchFunc func() (string, error)) (string, error) {
	val, err := rc.Client.Get(ctx, rc.WrapKey(key)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			val, err = fetchFunc()
			if err != nil {
				logger.Log.Errorf("failed to process fetch function: %v", err)
				return "", ErrFetchFunction
			}

			err = rc.Client.Set(ctx, rc.WrapKey(key), val, ttl).Err()
			if err != nil {
				logger.Log.Errorf("failed to set cache: %v", err)
				return "", ErrSetCache
			}
		} else {
			logger.Log.Errorf("failed to get cache: %v", err)
			return "", ErrGetCache
		}
	}

	return val, nil
}

func (rc *RedisConnection) IncrementCache(ctx context.Context, key string, increasedBy int64) (int64, error) {
	val, err := rc.Client.IncrBy(ctx, rc.WrapKey(key), increasedBy).Result()
	if err != nil {
		logger.Log.Errorf("failed to Increment cache: %v", err)
		return 0, ErrIncrementCache
	}

	return val, nil
}

func (rc *RedisConnection) DecrementCache(ctx context.Context, key string, decreasedBy int64) (int64, error) {
	val, err := rc.Client.DecrBy(ctx, rc.WrapKey(key), decreasedBy).Result()
	if err != nil {
		logger.Log.Errorf("failed to Decrement cache: %v", err)
		return 0, ErrDecrementCache
	}

	return val, nil
}
