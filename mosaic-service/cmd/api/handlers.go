package main

import (
	"encoding/json"
	"net/http"

	"github.com/ChrisShia/mosaic/cmd/internal"
)

func (app *App) createMosaicHandler(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		IP        string `json:"ip"`
		TileWidth int    `json:"tile_width"`
		Original  string `json:"original"`
	}

	decoder := json.NewDecoder(request.Body)

	err := decoder.Decode(&input)
	if err != nil {
		app.logger.PrintError(err, nil)
	}

	//TODO: this should get the index if it exists
	redisIndex := internal.NewRedisIndex(input.IP, redisIndexPrefix(input.IP), app.redisClient)

	originalImg, err := internal.Base64StringToImage(input.Original)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}
	b := NewMosaicBuilder(redisIndex, originalImg, input.TileWidth)

	mosaicImg, err := b.Mosaic()
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}

	base64StringImg, err := internal.ImageToBase64String(mosaicImg)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}

	var response struct {
		Error  bool   `json:"error"`
		Mosaic string `json:"mosaic"`
	}

	response.Mosaic = base64StringImg
	response.Error = false

	writer.Header().Set("Content-Type", "application/json")
	marshalledResponse, err := json.Marshal(response)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}

	_, err = writer.Write(marshalledResponse)
	if err != nil {
		app.logger.PrintError(err, nil)
	}
}
