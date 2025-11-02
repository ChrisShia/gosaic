package main

import (
	"context"
	"downloader/picsum"
	"encoding/json"
	"image"
	"io"
	"net/http"
	"os"

	"downloader/cmd/internal"
)

func (app *App) DownloadNRandomPicsFromPicSumHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: http.MaxBytesReader

	picSumRandomPicRequest := picsum.Random200300()

	var requestData struct {
		IP string `json:"ip"`
		N  int    `json:"n"`
	}

	dec := json.NewDecoder(r.Body)

	//dec.DisallowUnknownFields()

	err := dec.Decode(&requestData)
	if err != nil {
		app.logger.PrintError(err, map[string]string{
			"request":      r.URL.String(),
			"requestor_ip": requestData.IP,
		})
		return
	}

	//TODO: Ip address as a field since the request is essentially made from the broker(?)
	requestorIndexPrefix := app.redisIndexPrefix(requestData.IP)
	redisIndex := internal.NewRedisIndex(requestData.IP, requestorIndexPrefix, app.cfg.Redis.Client)
	redisIndex.FTCREATE()

	d := internal.NewDownloader(app.saveToRedis, app.logger)
	d.DownloadN(app.cfg.Nats.Client, requestData.IP, requestorIndexPrefix, requestData.N, picSumRandomPicRequest)
}

func (app *App) saveToFile(from io.Reader) error {
	f, err := os.CreateTemp(targetDirectory, "*.jpg")
	if err != nil {
		return err
	}

	_, err = io.Copy(f, from)
	if err != nil {
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}

	return nil
}

func (app *App) saveToRedis(ip, key string, from io.Reader) {
	img, err := app.Image(from)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}

	_ = internal.SaveToRedis(img, app.cfg.Redis.Client, ip, key, internal.ImageAverageRGB, context.Background())
}

func (app *App) Image(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	return img, nil
}
