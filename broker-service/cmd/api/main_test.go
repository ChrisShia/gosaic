package main

import "testing"

var mockApp = &App{
	services: map[string]string{
		"downloader": "http://localhost:4002/pic.sum/random/download",
		"mosaic":     "http://localhost:4001/create",
	},
}

func Test_randomTilesMosaicCreateRequest(t *testing.T) {
	err := mockApp.randomTilesMosaicCreateRequest("127.0.0.1", 10)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_downloadRandomNRequest(t *testing.T) {
	err := mockApp.downloadRandomNRequest("127.0.0.1", 25)
	if err != nil {
		t.Fatal(err)
	}
}
