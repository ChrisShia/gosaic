package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

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
