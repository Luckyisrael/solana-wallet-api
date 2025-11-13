package redis

import (
    "context"
    "time"
    "encoding/json"

    "github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

func NewClient(addr, password string, db int) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &Client{rdb}
}

func (c *Client) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, data, ttl).Err()
}

func (c *Client) GetJSON(ctx context.Context, key string, dest any) error {
	data, err := c.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
