package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (app *App) createMosaicHandler(writer http.ResponseWriter, request *http.Request) {
	var payload struct {
		IP        string `json:"ip"`
		TileWidth int    `json:"tile_width"`
	}

	decoder := json.NewDecoder(request.Body)

	err := decoder.Decode(&payload)
	if err != nil {
		app.logger.PrintError(err, nil)
	}

	app.logger.PrintInfo("payload", map[string]string{
		"ip":         payload.IP,
		"tile_width": strconv.Itoa(payload.TileWidth),
	})
}
