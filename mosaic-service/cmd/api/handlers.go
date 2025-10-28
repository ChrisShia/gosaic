package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"strings"

	"github.com/ChrisShia/mosaic/cmd/internal"
)

func (app *App) createMosaicHandler(writer http.ResponseWriter, request *http.Request) {
	var payload struct {
		IP        string `json:"ip"`
		TileWidth int    `json:"tile_width"`
		Image     string `json:"image"`
	}

	decoder := json.NewDecoder(request.Body)

	err := decoder.Decode(&payload)
	if err != nil {
		app.logger.PrintError(err, nil)
	}

	redisIndex := internal.NewRedisIndex(payload.IP, redisIndexPrefix(payload.IP), app.cfg.Redis.Client)

	imgReader := strings.NewReader(payload.Image)

	originalImg, _, _ := image.Decode(imgReader)
	b := NewMosaicBuilder(redisIndex, originalImg, payload.TileWidth)

	mosaicImg, err := b.Mosaic()
	if err != nil {
		app.logger.PrintError(err, nil)
	}

	err = png.Encode(writer, mosaicImg)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}
}

func redisIndexPrefix(ip string) string {
	return fmt.Sprintf("img:%s", ip)
}
