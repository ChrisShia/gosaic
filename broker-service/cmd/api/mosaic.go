package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type MosaicPayload struct {
	IP        string `json:"ip"`
	Original  string `json:"original,omitempty"`
	TileWidth int    `json:"tile_width,omitempty"`
}

func (app *App) randomTilesMosaicCreateRequest(ip string, original string, tileWidth int) (*string, error) {
	var mp = MosaicPayload{
		IP:        ip,
		Original:  original,
		TileWidth: tileWidth,
	}

	jsonData, err := json.Marshal(&mp)
	if err != nil {
		return nil, err
	}

	logServiceURL := app.service("mosaic")

	request, err := http.NewRequest(http.MethodPost, logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	res, err := client.Do(request)
	if err != nil {
		//NOTE: EOF error
		return nil, err
	}

	var mosaicServiceResponse struct {
		Error  bool   `json:"error,omitempty"`
		Mosaic string `json:"mosaic"`
	}

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&mosaicServiceResponse)
	if err != nil {
		app.logger.PrintError(err, nil)
		return nil, err
	}

	return &mosaicServiceResponse.Mosaic, nil
}
