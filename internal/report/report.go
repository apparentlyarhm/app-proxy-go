package report

import (
	"context"
	"time"

	"github.com/apparentlyarhm/app-proxy-go/config"
	"github.com/redis/go-redis/v9"
)

// a wrapper around a wrapper, mad.
// Well this was done to respect the already present structure.
// TODO: perhaps refactor?
type Client struct {
	client *redis.Client
}

func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Since this is an actual client, good to return err, especially if the connection fails - this will be handled at main
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return &Client{client: rdb}, nil
}

func (c *Client) Close() error {
	// adding it now we will see if we need to do it or not
	return c.client.Close()
}
