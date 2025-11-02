package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"os"
	"testing"
)

var mockApp = &App{
	services: map[string]string{
		"downloader": "http://localhost:4002/pic.sum/random/download",
		"mosaic":     "http://localhost:4001/create",
	},
}

func Test_randomTilesMosaicCreateRequest(t *testing.T) {
	img, err := PngDecode("../../../test_image_700.png")

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		t.Fatal(err)
		return
	}

	originalStr := base64.StdEncoding.EncodeToString(buf.Bytes())

	mosaicImgStr, err := mockApp.randomTilesMosaicCreateRequest("127.0.0.1", originalStr, 20)
	if err != nil {
		t.Fatal(err)
	}

	resFile, err := os.Create("mosaic.png")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer resFile.Close()

	resImage, err := base64StringToImage(*mosaicImgStr)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = png.Encode(resFile, resImage)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func Test_downloadRandomNRequest(t *testing.T) {
	err := mockApp.downloadRandomNRequest("127.0.0.1", 800)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Image(t *testing.T) {
	file, err := os.Open("../../../test_image_700.png")
	if err != nil {
		t.Fatal(err)
	}

	_, err = Image(file)
	if err != nil {
		t.Fatal(err)
	}

}

func PngDecode(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}
