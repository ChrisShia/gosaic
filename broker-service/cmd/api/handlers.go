package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"net"
	"net/http"
)

func (app *App) mosaicHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Original  string `json:"original"`
		TileWidth int    `json:"tile_width,omitempty"`
	}

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

	originalImg, err := base64StringToImage(payload.Original)
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

	w.Header().Set("Content-Type", "application/json")

	mosaicStr, err := app.randomTilesMosaicCreateRequest(host, payload.Original, payload.TileWidth)
	if err != nil {
		app.logger.PrintError(err, nil)
		//TODO: specialize error handling

		app.errorResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"mosaic": mosaicStr}, nil)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}
}

func Image(r io.Reader) (image.Image, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func base64StringToImage(str string) (image.Image, error) {
	decodedLen := base64.StdEncoding.DecodedLen(len(str))
	p := make([]byte, decodedLen)

	_, err := base64.StdEncoding.Decode(p, []byte(str))
	if err != nil {
		return nil, err
	}

	return Image(bytes.NewBuffer(p))
}
