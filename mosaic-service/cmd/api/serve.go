package main

import "fmt"

func (app *App) LogStartUp() {
	app.logger.PrintInfo("Listening on server:", map[string]string{
		"Port": fmt.Sprintf(":%d", app.cfg.Port),
	})
}

func (app *App) LogShutdown() {

}

func (app *App) PrintInfo(msg string, properties map[string]string) {
	app.logger.PrintInfo(msg, properties)
}

func (app *App) Write(p []byte) (int, error) {
	return app.logger.Write(p)
}
