package main

import (
	"net/http"
)

func (app *App) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/mosaic", app.mosaicHandler)

	return mux
}
