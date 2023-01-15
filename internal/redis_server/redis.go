package redis_server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/geolocket/batch_redis/internal/momentType"
	"github.com/geolocket/batch_redis/internal/postgres"
	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	"log"
	"os"
	"time"
)

type redisAPI interface {
	Pipeline() redis.Pipeliner
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Close() error
}

func redisNewClient() redisAPI {
	env := os.Getenv("ENV")
	var c redisAPI
	if env == "PROD" {
		c = redis.NewClusterClient(&redis.ClusterOptions{

			Addrs:    []string{os.Getenv("REDIS_ADDR")},
			Username: os.Getenv("REDIS_USERNAME"),
			Password: os.Getenv("REDIS_PASSWORD"),
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		})
	} else if env == "DEV" {
		c = redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_ADDR"),
			Username: os.Getenv("REDIS_USERNAME"),
			Password: os.Getenv("REDIS_PASSWORD"),
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				//Certificates: []tls.Certificate{cert}
			},
		})
	} else {
		c = redis.NewClient(&redis.Options{
			Addr: "host.docker.internal:6379",
		})
	}

	return c
}

func BatchRedis(now time.Time) {
	c := redisNewClient()

	p := c.Pipeline()
	ctx := context.Background()

	l := 0
	momentMap := make(map[string]momentType.RedisMomentBody, 100000)

	for _, v := range momentType.MomentQueue {
		if v.RegisterTime.After(now) {
			break
		}
		l += 1
		u, err := json.Marshal(v.RedisMomentBody)
		if err != nil {
			log.Println(errors.WithStack(err))
			return
		}

		err = p.Set(ctx, v.UserId, u, time.Minute*30).Err()
		if err != nil {
			log.Println(errors.WithStack(err))
			return
		}
		momentMap[v.UserId] = v.RedisMomentBody

	}
	err := postgres.UpdateMoment(momentMap)
	if err != nil {
		log.Printf("%+v", err)
	}
	if l > 0 {
		fmt.Println(l, "数", momentType.MomentQueue)
	}
	momentType.MomentQueue = momentType.MomentQueue[l:]
	_, err = p.Exec(ctx)
	if err != nil {
		log.Println(errors.WithStack(err))
	}

}

func ReadRedis(key string) (string, error) {
	c := redisNewClient()
	ctx := context.Background()

	// Scan all fields into the model
	v, err := c.Get(ctx, key).Result()

	if err != nil {
		return "", errors.WithStack(err)
	}
	fmt.Println(v)

	return v, nil
}
