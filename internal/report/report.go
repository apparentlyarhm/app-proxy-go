package report

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/apparentlyarhm/app-proxy-go/config"
	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidSignature = errors.New("invalid hmac signature")
	ErrMalformedPayload = errors.New("malformed payload")
	ErrTooFast          = errors.New("view recorded too fast")
	ErrExpired          = errors.New("view ticket expired")
	ErrDuplicate        = errors.New("view already recorded")
)

// a wrapper around a wrapper, mad.
// Well this was done to respect the already present structure.
// TODO: perhaps refactor?
type Client struct {
	actualRedisClient *redis.Client
	cryptocfg         config.CryptoConfig
}

func NewClient(cfg config.RedisConfig, cryptoCfg config.CryptoConfig) (*Client, error) {
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
	return &Client{actualRedisClient: rdb, cryptocfg: cryptoCfg}, nil
}

func (c *Client) Close() error {
	// adding it now we will see if we need to do it or not
	return c.actualRedisClient.Close()
}

func (c *Client) PutSystemReport(ctx context.Context, req SystemInfo) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Println("[REPORT SERV] :: marshaling error:", err)
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
		log.Printf("ERROR WHILE READING %v", e.Error())
		return nil, e

	}

	var res SystemInfo

	err := json.Unmarshal([]byte(rawData), &res)
	if err != nil {
		log.Println("[REPORT SERV] :: UNmarshaling error:", err)
		return nil, err
	}

	return res, nil

}

func (c *Client) DeleteReport(ctx context.Context) error {

	res, e := c.actualRedisClient.Del(ctx, "si").Result()
	if e != nil {
		log.Printf("[REPORT SERV] :: error while deleting %v", e.Error())
		return e
	}

	log.Printf("[REPORT SERV] :: deleted %v entry(s)", res)
	return nil
}

// we take the data, we hash it with our secret and then we compare it to the signature sent by the client. since the incoming request
// is from a known server, and with shared secrets, we can be sure that the data is not tampered with if the signature is valid.
func (c *Client) verifySignature(base64Payload string, signature string) bool {
	h := hmac.New(sha256.New, []byte(c.cryptocfg.HMACSecret))
	h.Write([]byte(base64Payload))
	expectedMac := hex.EncodeToString(h.Sum(nil))

	// Safely compare to prevent timing attacks
	return hmac.Equal([]byte(signature), []byte(expectedMac))
}

/*
the core business logic of:
1. verifying the signature
2. decoding the payload
3. enforcing business rules (time based)
4. using Redis SETNX to prevent duplicates
5. recording a "view"
*/
func (c *Client) RecordBlogView(ctx context.Context, req ViewRequest) error {

	if !c.verifySignature(req.Payload, req.Signature) {
		log.Printf("[REPORT SERV] :: Invalid signature attempt")
		return ErrInvalidSignature
	}

	jsonBytes, err := base64.StdEncoding.DecodeString(req.Payload)
	if err != nil {
		return ErrMalformedPayload
	}

	var payload ViewPayload
	if err := json.Unmarshal(jsonBytes, &payload); err != nil {
		return ErrMalformedPayload
	}

	now := time.Now().Unix()
	if now-payload.IssuedAt < 3 {
		return ErrTooFast
	}
	if now-payload.IssuedAt > 3600 {
		return ErrExpired
	}

	// TODO: keep a watch on this approach
	// we will accumulate counts in redis and flush externally.
	cooldownKey := fmt.Sprintf("view_cooldown:%s:%s", payload.Slug, payload.ViewerID)

	isNew, err := c.actualRedisClient.SetNX(ctx, cooldownKey, "1", 24*time.Hour).Result()
	if err != nil {
		return fmt.Errorf("redis setnx error: %w", err)
	}
	if !isNew {
		// The script is looping, or the user refreshed. Ignore silently or throw error.
		return ErrDuplicate
	}

	// This creates/updates a hash where Field = Slug, Value = Count
	// pbv == pending_blog_views
	// this count will be processed.
	err = c.actualRedisClient.HIncrBy(ctx, "pbv", payload.Slug, 1).Err()
	if err != nil {
		return fmt.Errorf("redis hincrby error: %w", err)
	}

	return nil
}
