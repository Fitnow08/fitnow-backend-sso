package connect

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func RedisConnect(ctx context.Context, Host, Port, Password, Env string, dbnumber, retries int) (*redis.Client, error) {
	url := BuildRedisURL("default", Password, Host, Port, 0)
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	opt.MaxRetries = retries
	var rdb *redis.Client
	switch Env {
	case "development":
		rdb = redis.NewClient(&redis.Options{
			Addr:       "localhost:6379",
			MaxRetries: 5,
			DB:         0,
		})
	case "production":
		rdb = redis.NewClient(opt)
	}

	ping, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	fmt.Println("Redis connect success:", ping)
	return rdb, nil
}
func BuildRedisURL(username, password, host, port string, db int) string {
	if username != "" && password != "" {
		return fmt.Sprintf("redis://%s:%s@%s:%s/%d", username, password, host, port, db)
	} else if password != "" {
		return fmt.Sprintf("redis://:%s@%s:%s/%d", password, host, port, db)
	}
	return fmt.Sprintf("redis://%s:%s/%d", host, port, db)
}
