package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(address string, db int) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr: address,
		DB:   db,
	})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	return &Redis{client: client}, nil
}

func (c *Redis) Close() error {
	if c.client != nil {
		err := c.client.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
