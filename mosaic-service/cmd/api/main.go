package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/ChrisShia/jsonlog"
	"github.com/ChrisShia/serve"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Port int

	Redis struct {
		Addr   string
		Client *redis.Client
	}
}

type App struct {
	logger *jsonlog.Logger
	cfg    Config
}

func main() {
	cfg := Config{}
	cfg.flags()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	app := &App{
		logger: logger,
		cfg:    cfg,
	}

	//TODO: might have to connect to NATS

	redisClose, err := app.connectToRedis(cfg)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}
	defer redisClose()

	err = serve.ListenAndServe(app, app.cfg.Port)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}
}

func (app *App) connectToRedis(cfg Config) (func(), error) {
	counts := 0
	for {
		client, err := establishRedisConnAndPing(cfg)
		if err != nil {
			counts++
		} else {
			app.cfg.Redis.Client = client
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

func (app *App) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/create", app.createMosaicHandler)

	return mux
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
