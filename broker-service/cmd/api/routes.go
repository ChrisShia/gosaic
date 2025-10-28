package main

import (
	"bytes"
	"encoding/json"
	"image"
	"io"
	"net"
	"net/http"
)

func (app *App) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/mosaic", app.mosaicHandler)

	return mux
}

type MosaicPayload struct {
	IP string `json:"ip"`
	//Original []byte `json:"original,omitempty"`
	TileWidth int `json:"tile_width,omitempty"`
}

type BrokerPayload struct {
	Original  []byte `json:"original"`
	TileWidth int    `json:"tile_width"`
}

func (app *App) mosaicHandler(w http.ResponseWriter, r *http.Request) {
	var payload BrokerPayload

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&payload)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		app.badRequestResponse(w, r, err)
		return
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		//TODO: use json helper for errors...along with jsonResponse struct{}
	}

	originalImg, err := Image(bytes.NewBuffer(payload.Original))
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}

	//NOTE: approximate
	dx := originalImg.Bounds().Dx() / payload.TileWidth
	dy := originalImg.Bounds().Dy() / payload.TileWidth
	tilesNeeded := dx * dy

	err = app.downloadRandomNRequest(host, tilesNeeded)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}

	err = app.randomTilesMosaicCreateRequest(host, payload.TileWidth)
	if err != nil {
		//TODO: specialize error handling
		return
	}
}

func Image(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (app *App) randomTilesMosaicCreateRequest(ip string, tileWidth int) error {
	var mp = MosaicPayload{
		IP:        ip,
		TileWidth: tileWidth,
	}

	jsonData, err := json.Marshal(&mp)
	if err != nil {
		return err
	}

	logServiceURL := app.service("mosaic")

	request, err := http.NewRequest(http.MethodPost, logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}

type DownloadPayload struct {
	IP string `json:"ip"`
	N  int    `json:"n"`
}

func (app *App) downloadRandomNRequest(host string, n int) error {
	var dp = DownloadPayload{
		IP: host,
		N:  n,
	}

	jsonData, err := json.Marshal(&dp)
	if err != nil {
		return err
	}

	logServiceURL := app.service("downloader")

	request, err := http.NewRequest(http.MethodPost, logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := http.Client{}

	//TODO: should treat response appropriately...maybe errors occurred in the process
	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
