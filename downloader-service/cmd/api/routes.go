package main

import "net/http"

func (app *App) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/pic.sum/random/download", app.DownloadNRandomPicsFromPicSumHandler)

	return mux
}
