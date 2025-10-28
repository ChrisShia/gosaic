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
	redisIndex := internal.NewRedisIndex(requestData.IP, app.redisIndexPrefix(requestData.IP), app.cfg.Redis.Client)
	redisIndex.FTCREATE()

	d := internal.NewDownloader(app.saveToRedis, app.logger)
	d.DownloadN(app.cfg.Nats.Client, requestData.IP, requestData.N, picSumRandomPicRequest)
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

func (app *App) saveToRedis(ip string, from io.Reader) {
	img, err := app.Image(from)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}

	indexPrefix := "img"

	_ = internal.SaveToRedis(img, app.cfg.Redis.Client, ip, indexPrefix, averageColor, context.Background())
}

func (app *App) Image(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func averageColor(img image.Image) ([3]float64, error) {
	bounds := img.Bounds()
	r, g, b := 0.0, 0.0, 0.0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r1, g1, b1, _ := img.At(x, y).RGBA()
			r, g, b = r+float64(r1), g+float64(g1), b+float64(b1)
		}
	}

	totalPixels := float64(bounds.Max.X * bounds.Max.Y)
	return [3]float64{r / totalPixels, g / totalPixels, b / totalPixels}, nil
}
