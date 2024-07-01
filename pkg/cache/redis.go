package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
)

type Redis struct {
	Client *redis.Client
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

	return &Redis{Client: client}, nil
}

func (c *Redis) Close() error {
	if c.Client != nil {
		err := c.Client.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Redis) LRange(key string) ([]string, error) {
	return c.Client.LRange(context.Background(), key, 0, -1).Result()
}

func (c *Redis) RPush(key string, values ...interface{}) error {
	return c.Client.RPush(context.Background(), key, values...).Err()
}
func (c *Redis) Del(pattern string) error {
	var keys []string
	var cursor uint64
	for {
		result, nextCursor, err := c.Client.Scan(context.Background(), cursor, pattern, 1000).Result()
		if err != nil {
			log.Printf("Error scanning cached messages: %v\n", err)
			break
		}
		keys = append(keys, result...)
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	return c.Client.Del(context.Background(), keys...).Err()
}
