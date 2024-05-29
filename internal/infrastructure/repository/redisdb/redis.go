package redis

import (
	"context"
	"encoding/json"
	"time"


	redis "univer/internal/pkg/storage"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Del(ctx context.Context, key string) error
}

func NewCache(rdb *redis.RedisDB) *cache {
	return &cache{
		rdb: rdb,
	}
}

type cache struct {
	rdb *redis.RedisDB
}

func (c *cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	
	byteData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = c.rdb.Client.Set(ctx, key, string(byteData), expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c *cache) Get(ctx context.Context, key string) ([]byte, error) {
	

	data, err := c.rdb.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	
	return []byte(data), nil
}

func (c *cache) Del(ctx context.Context, key string) error {
	

	err := c.rdb.Client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}
