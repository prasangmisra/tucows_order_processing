package redisconn

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

func Connect() (*redis.Client, error) {
	redisPort := os.Getenv("REDIS_PORT")

	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:" + redisPort,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Redis: %v", err)
	}

	return rdb, nil
}

func Subscribe(rdb *redis.Client, channelName string) <-chan *redis.Message {
	subscriber := rdb.Subscribe(context.Background(), channelName)
	return subscriber.Channel()
}
