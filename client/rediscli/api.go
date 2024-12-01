package redis_go

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisValueTypes interface {
	int | int64 | string | bool | []byte
}

func Get[T RedisValueTypes](ctx context.Context, key string) (val T, isNotExist bool, err error) {
	var result any
	switch any(val).(type) {
	case string:
		result, err = redisClient.Get(ctx, key).Result()
	case int:
		result, err = redisClient.Get(ctx, key).Int()
	case int64:
		result, err = redisClient.Get(ctx, key).Int64()
	case bool:
		result, err = redisClient.Get(ctx, key).Bool()
	case []byte:
		result, err = redisClient.Get(ctx, key).Bytes()
	default:
		err = errors.New("redis get with unsupported type")
	}
	if err != nil && errors.Is(err, redis.Nil) {
		isNotExist = true
		err = nil
	}
	val, ok := result.(T)
	if !ok {
		err = fmt.Errorf("failed to cast result to type %T", val)
	}
	return
}

func SetEx(ctx context.Context, key string, val any, exp time.Duration) (err error) {
	err = redisClient.Set(ctx, key, val, exp).Err()
	return
}

func DelEx(ctx context.Context, key string) (err error) {
	err = redisClient.Del(ctx, key).Err()
	return
}

func HashSet(ctx context.Context, key string, val map[string]any, expiration time.Duration) (err error) {
	// 使用 HSet 命令设置 Hash 值
	err = redisClient.HMSet(ctx, key, val).Err()
	if err != nil {
		return
	}
	// 设置过期时间
	if expiration > 0 {
		err = redisClient.Expire(ctx, key, expiration).Err()
	}
	return
}

func HashGetAll(ctx context.Context, key string) (result map[string]string, isNotExist bool, err error) {
	// 使用 HGetAll 命令获取 Hash 的所有字段和值
	result, err = redisClient.HGetAll(ctx, key).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		isNotExist = true
		err = nil
	}
	return
}

func HashUpdate(ctx context.Context, key string, val ...any) (err error) {
	err = redisClient.HSet(ctx, key, val).Err()
	return
}

func HashDel(ctx context.Context, key string, val string) (err error) {
	err = redisClient.HDel(ctx, key, val).Err()
	return
}

func HashGetString(ctx context.Context, key, field string) (result string, isNotExist bool, err error) {
	result, err = redisClient.HGet(ctx, key, field).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		isNotExist = true
		err = nil
	}
	return
}

func Subscribe(ctx context.Context, key string) (subscribe *redis.PubSub) {
	subscribe = redisClient.Subscribe(ctx, key)
	return
}

// GetKeyEventExpired 需要修改redis.conf配置项，启用notify-keyspace-events = "EX"
func GetKeyEventExpired(db int) string {
	return fmt.Sprintf("__keyevent@%d__:expired", db)
}