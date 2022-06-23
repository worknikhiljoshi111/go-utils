package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

type mockRedisData struct {
	key string
	val string
}

func TestNewConn(t *testing.T) {

	var ctx = context.Background()
	var conn RedisConn
	err := conn.NewConn(ctx)
	assert.EqualError(t, err, "host path is missing")
}

func TestGet(t *testing.T) {

	var ctx = context.Background()
	var conn RedisConn
	db, mock := redismock.NewClientMock()
	conn.client = db
	mockr := mockRedisData{key: "news_redis_cache_123456", val: "s"}
	mock.ExpectGet(mockr.key).SetVal(mockr.val)
	res, err := conn.Get(ctx, mockr.key)
	assert.NoError(t, err)
	assert.Equal(t, mockr.val, res)

}

func TestSet(t *testing.T) {
	var conn RedisConn
	var mock redismock.ClientMock
	var ctx = context.Background()
	mockr := mockRedisData{key: "news_redis_cache_123456", val: "s"}
	conn.client, mock = redismock.NewClientMock()
	mock.Regexp().ExpectSet(mockr.key, `[a-z]+`, 30*time.Minute).SetErr(errors.New("FAIL"))
	err := conn.Set(ctx, mockr.key, mockr.val, 30*time.Minute)

	assert.Equal(t, "FAIL", err.Error())

}
