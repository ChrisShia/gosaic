package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ChrisShia/jsonlog"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

const targetDirectory = "Downloads"

type Config struct {
	Port int
	Fs   bool
	Nats struct {
		Embedded bool
		Url      string
		Client   *nats.Conn
	}
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

	redisClose, err := app.connectToRedis(cfg)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}
	defer redisClose()

	natsClose, err := app.connectToNats(cfg)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}
	defer natsClose()

	addr := fmt.Sprintf(":%d", app.cfg.Port)
	err = http.ListenAndServe(addr, app.routes())
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}
}

func (app *App) connectToNats(cfg Config) (func(), error) {
	if cfg.Nats.Embedded {
		//app.logger.PrintInfo("Starting Nats Server...", nil)
		//natsServer := NatsServer()
		//app..Start()
		//d.logger.PrintInfo("Nats Started!", nil)
		//defer func() {
		//	d.logger.PrintInfo("Stopping Nats Server...", nil)
		//	d.NATS.Shutdown()
		//	d.NATS.WaitForShutdown()
		//}()
		return nil, nil
	}

	client, err := nats.Connect(cfg.Nats.Url)
	if err != nil {
		return nil, err
	}

	app.logger.PrintInfo("Connected to NATS", map[string]string{
		"nats_url": cfg.Nats.Url,
	})

	app.cfg.Nats.Client = client

	return func() { client.Close() }, nil
}

func createTargetDir() {
	if _, err := os.Stat(targetDirectory); err != nil {
		switch {
		case os.IsNotExist(err):
			err = os.Mkdir(targetDirectory, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
		default:
			log.Fatal(err)
		}
	}
}

func (app *App) connectToRedis(cfg Config) (func(), error) {
	counts := 0
	for {
		client, err := establishRedisConnAndPing(cfg)
		if err != nil {
			counts++
		} else {
			app.logger.PrintInfo("Connected to redis", map[string]string{
				"redis_url": client.String(),
			})
			app.cfg.Redis.Client = client
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

func (app *App) redisFTCREATE(ip string) {
	//options := redis.FTCreateOptions{
	//	OnHash: true,
	//	Prefix: []any{"img"},
	//}
	//app.cfg.Redis.Client.FTCreate()

	_, err := app.cfg.Redis.Client.Do(context.Background(),
		"FT.CREATE",
		"img_idx",
		"ON",
		"HASH",
		"PREFIX", "1", app.redisIndexPrefix(ip),
		"SCHEMA", "avg_color", "VECTOR", "HNSW", "6",
		"TYPE", "FLOAT64",
		"DIM", "3",
		"DISTANCE_METRIC", "L2",
	).Result()
	if err != nil {
		// ignore error if index already exists
	}
}

func (app *App) redisIndexPrefix(ip string) string {
	return fmt.Sprintf("img:%s:", ip)
}
