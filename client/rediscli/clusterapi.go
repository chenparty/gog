package rediscli

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// ClusterGet 获取key的值
func ClusterGet[T RedisValueTypes](ctx context.Context, key string) (val T, isNotExist bool, err error) {

	var result any
	switch any(val).(type) {
	case string:
		result, err = redisClusterClient.Get(ctx, key).Result()
	case int:
		result, err = redisClusterClient.Get(ctx, key).Int()
	case int64:
		result, err = redisClusterClient.Get(ctx, key).Int64()
	case bool:
		result, err = redisClusterClient.Get(ctx, key).Bool()
	case []byte:
		result, err = redisClusterClient.Get(ctx, key).Bytes()
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

// ClusterSetEx 设置key和值，并设置过期时间
func ClusterSetEx(ctx context.Context, key string, val any, exp time.Duration) (err error) {
	err = redisClusterClient.Set(ctx, key, val, exp).Err()
	return
}

// ClusterDel 删除key
func ClusterDel(ctx context.Context, key string) (err error) {
	err = redisClusterClient.Del(ctx, key).Err()
	return
}

// ClusterHashSet 设置Hash
func ClusterHashSet(ctx context.Context, key string, val map[string]any, expiration time.Duration) (err error) {
	// 使用 HSet 命令设置 Hash 值
	err = redisClusterClient.HMSet(ctx, key, val).Err()
	if err != nil {
		return
	}
	// 设置过期时间
	if expiration > 0 {
		err = redisClusterClient.Expire(ctx, key, expiration).Err()
	}
	return
}

// ClusterHashUpdate 更新Hash字段
func ClusterHashUpdate(ctx context.Context, key string, val ...any) (err error) {
	err = redisClusterClient.HSet(ctx, key, val).Err()
	return
}

// ClusterHashDel 删除Hash字段
func ClusterHashDel(ctx context.Context, key string, val string) (err error) {
	err = redisClusterClient.HDel(ctx, key, val).Err()
	return
}

// ClusterHashGetString 获取Hash字段的值
func ClusterHashGetString(ctx context.Context, key, field string) (result string, isNotExist bool, err error) {
	result, err = redisClusterClient.HGet(ctx, key, field).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		isNotExist = true
		err = nil
	}
	return
}

// ClusterHashGetAll 获取Hash的所有字段和值
func ClusterHashGetAll(ctx context.Context, key string) (result map[string]string, isNotExist bool, err error) {
	// 使用 HGetAll 命令获取 Hash 的所有字段和值
	result, err = redisClusterClient.HGetAll(ctx, key).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		isNotExist = true
		err = nil
	}
	return
}

// ClusterSubscribe 订阅一个key
func ClusterSubscribe(ctx context.Context, key string) (subscribe *redis.PubSub) {
	subscribe = redisClusterClient.Subscribe(ctx, key)
	return
}

// ClusterGetKeyEventExpired 需要修改redis.conf配置项，启用notify-keyspace-events = "EX"
func ClusterGetKeyEventExpired(db int) string {
	return fmt.Sprintf("__keyevent@%d__:expired", db)
}

// ClusterIsRedisNilErr 判断是否为redis.Nil
func ClusterIsRedisNilErr(err error) bool {
	return errors.Is(err, redis.Nil)
}
