package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ChrisShia/jsonlog"
	"github.com/ChrisShia/serve"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Port int

	Redis struct {
		Addr string
	}

	mode mode
}

type mode uint8

const (
	normal mode = 1 << iota
	test
)

type App struct {
	logger      *jsonlog.Logger
	cfg         Config
	redisClient *redis.Client
}

func main() {
	cfg := Config{}
	cfg.flags()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	app := &App{
		logger: logger,
		cfg:    cfg,
	}

	closerFunc, err := app.setupTileImageRepository()
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}
	defer closerFunc()

	err = serve.ListenAndServe(app, app.cfg.Port)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}
}

func (app *App) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/create", app.createMosaicHandler)

	return mux
}

func redisIndexPrefix(ip string) string {
	return fmt.Sprintf("img:%s", ip)
}
