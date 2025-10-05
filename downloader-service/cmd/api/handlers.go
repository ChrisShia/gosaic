package main

import (
	"bytes"
	"context"
	"downloader/picsum"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"math"
	"net/http"
	"os"

	"downloader/cmd/internal"
)

func (app *App) DownloadNRandomPicsFromPicSumHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: http.MaxBytesReader

	d := internal.NewDownloader(app.saveToRedis, app.logger)

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

	app.redisFTCREATE(requestData.IP)

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

var ctx_ = context.Background()

func (app *App) saveToRedis(ip string, from io.Reader) {
	img, err := app.Image(from)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}

	ac := averageColor(img)
	avColorBinary := make([]byte, 8*len(ac))
	for i, f := range ac {
		binary.LittleEndian.PutUint64(avColorBinary[i*8:], math.Float64bits(f))
	}

	imgBuf := new(bytes.Buffer)
	err = jpeg.Encode(imgBuf, img, nil)
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}

	imgString := base64.StdEncoding.EncodeToString(imgBuf.Bytes())

	//data := struct {
	//	Img          string  `redis:"img"`
	//	AverageRed   float64 `redis:"average_red"`
	//	AverageGreen float64 `redis:"average_green"`
	//	AverageBlue  float64 `redis:"average_blue"`
	//}{
	//	Img:          imgString,
	//	AverageRed:   ac[0],
	//	AverageGreen: ac[1],
	//	AverageBlue:  ac[2],
	//}

	data := struct {
		Img          string `redis:"img"`
		AverageColor []byte `redis:"average"`
	}{
		Img:          imgString,
		AverageColor: avColorBinary,
	}

	app.logger.PrintInfo(imgString, nil)

	//NOTE:
	//create another entry in redis to keep track of the counter assigned
	//to a specific ip.
	counterKey := ip + ":counter"

	id, err := app.cfg.Redis.Client.Incr(ctx_, counterKey).Result()
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}

	key := fmt.Sprintf("%s:%d", app.redisIndexPrefix(ip), id)

	if err = app.cfg.Redis.Client.HSet(ctx_, key, data).Err(); err != nil {
		app.logger.PrintError(err, map[string]string{})
	}
}

func (app *App) Image(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func averageColor(img image.Image) [3]float64 {
	bounds := img.Bounds()
	r, g, b := 0.0, 0.0, 0.0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r1, g1, b1, _ := img.At(x, y).RGBA()
			r, g, b = r+float64(r1), g+float64(g1), b+float64(b1)
		}
	}
	totalPixels := float64(bounds.Max.X * bounds.Max.Y)
	return [3]float64{r / totalPixels, g / totalPixels, b / totalPixels}
}
