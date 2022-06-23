package redis

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisConn struct {
	client *redis.Client
}

// NewConn creates the new redis connection
// - returns error on failures
func (r RedisConn) NewConn(ctx context.Context) error {
	var h, p string
	var ok bool
	if h, ok = os.LookupEnv("REDIS_HOST"); !ok {
		return errors.New("host path is missing")
	}

	p, _ = os.LookupEnv("REDIS_PORT")

	// set Default
	r.client = redis.NewClient(&redis.Options{
		Addr:        h + ":" + p,
		IdleTimeout: 4 * time.Hour,
	})

	if r.client == nil {
		return errors.New("redis connection failed")
	}

	if _, err := r.client.Ping(ctx).Result(); err != nil {

		return err
	}

	return nil
}

// Set sets redis key
// takes key, value and expiry returns error if fails or nil on success
func (r RedisConn) Set(ctx context.Context, k string, d string, e time.Duration) error {

	if k == "" {
		return errors.New("key parameter missing")
	}

	if d == "" {
		return errors.New("value parameter missing")
	}
	return r.client.Set(ctx, k, d, e).Err()
}

// getRedisKey
// takes key as input and returns string as output or error
func (r RedisConn) Get(ctx context.Context, k string) (string, error) {

	if k == "" {
		return "", errors.New("key parameter missing")
	}
	return r.client.Get(ctx, k).Result()

}

func (r RedisConn) Expire(ctx context.Context, k string, t time.Duration) error {

	if k == "" {
		return errors.New("key parameter missing")
	}
	return r.client.Expire(ctx, k, t).Err()
}
