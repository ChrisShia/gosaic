package main

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func (app *App) setupTileImageRepository() (func(), error) {
	redisClose, err := app.connectToRedis(app.cfg)
	if err != nil {
		app.logger.PrintError(err, nil)
		return nil, err
	}

	return redisClose, nil
}

func (app *App) connectToRedis(cfg Config) (func(), error) {
	counts := 0
	for {
		client, err := establishRedisConnAndPing(cfg)
		if err != nil {
			app.logger.PrintError(err, nil)
			counts++
		} else {
			app.redisClient = client
			app.logger.PrintInfo("Connected to redis", map[string]string{
				"addr": cfg.Redis.Addr,
			})
			return func() { client.Close() }, nil
		}

		if counts > 5 {
			return nil, err
		}
	}
}

func establishRedisConnAndPing(cfg Config) (*redis.Client, error) {
	client, err := redisClient(cfg)
	if err != nil {
		return nil, err
	}

	timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	if err = client.Ping(timeout).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

func redisClient(cfg Config) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.Redis.Addr)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	return client, nil
}
