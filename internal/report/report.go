package report

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/apparentlyarhm/app-proxy-go/config"
	"github.com/redis/go-redis/v9"
)

// a wrapper around a wrapper, mad.
// Well this was done to respect the already present structure.
// TODO: perhaps refactor?
type Client struct {
	actualRedisClient *redis.Client
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

	log.Println("[REPORT SERV] :: redis pinged")
	return &Client{actualRedisClient: rdb}, nil
}

func (c *Client) Close() error {
	// adding it now we will see if we need to do it or not
	return c.actualRedisClient.Close()
}

func (c *Client) PutSystemReport(ctx context.Context, req SystemInfo) error {

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Println("[REPORT SERV] :: marshaling error:", err)
		return err
	}

	payload := string(jsonData)

	e := c.actualRedisClient.Set(ctx, "si", payload, 0).Err() // this key will be constant and will be read.
	if e != nil {
		log.Printf("[REPORT SERV] :: err -> %v ", e.Error())
		return e
	}
	log.Println("[REPORT SERV] :: data written to redis!")

	return nil
}

func (c *Client) GetSystemReport(ctx context.Context) (any, error) {

	rawData, e := c.actualRedisClient.Get(ctx, "si").Result() // any error while finding the key is InternalServerError to the client because the key is hardcoded here.
	if e != nil {
		fmt.Printf("ERROR WHILE READING %v", e.Error())
		return nil, e

	}

	var res SystemInfo

	err := json.Unmarshal([]byte(rawData), &res)
	if err != nil {
		fmt.Println("[REPORT SERV] :: UNmarshaling error:", err)
		return nil, err
	}

	return res, nil

}
