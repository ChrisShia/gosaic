package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/ChrisShia/jsonlog"
)

type Config struct {
	Port int
}

func (c *Config) port() string {
	return fmt.Sprintf(":%d", c.Port)
}

func (c *Config) flags() {
	flag.IntVar(&c.Port, "p", 80, "http server port")

	flag.Parse()
}

type App struct {
	logger   *jsonlog.Logger
	cfg      Config
	services map[string]string
}

func main() {

	cfg := Config{}

	cfg.flags()

	app := App{
		logger:   jsonlog.New(os.Stdout, jsonlog.LevelInfo),
		cfg:      cfg,
		services: services(),
	}

	server := http.Server{
		Addr:    app.cfg.port(),
		Handler: app.Routes(),
	}

	err := server.ListenAndServe()
	if err != nil {
		app.logger.PrintError(err, nil)
	}
}

func services() map[string]string {
	srv := make(map[string]string)
	srv["mosaic"] = "http://mosaic-service/create"
	srv["downloader"] = "http://downloader-service/pic.sum/random/download"
	return srv
}

func (app *App) service(name string) string {
	return app.services[name]
}
