package producer

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisClient struct {
	client   *redis.Client
	ctx      context.Context
	redisKey string
}

func NewRedisClient(host string, port int, password string) *RedisClient {
	addr := fmt.Sprintf("%s:%d", host, port)
	rdb := redis.NewClient(&redis.Options{
		Addr:        addr,
		Password:    password,
		DB:          0,
		DialTimeout: 2 * time.Second,
	})

	return &RedisClient{
		client:   rdb,
		ctx:      context.Background(),
		redisKey: "configs",
	}
}

func (r *RedisClient) AddConfig(config string) (bool, error) {
	res, err := r.client.SAdd(r.ctx, r.redisKey, config).Result()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

func (r *RedisClient) ConfigExists(config string) (bool, error) {
	return r.client.SIsMember(r.ctx, r.redisKey, config).Result()
}

func (r *RedisClient) ListAllConfigs() error {
	members, err := r.client.SMembers(r.ctx, r.redisKey).Result()
	if err != nil {
		return err
	}

	fmt.Println("Configs:")
	for _, m := range members {
		fmt.Println(m)
	}
	return nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
