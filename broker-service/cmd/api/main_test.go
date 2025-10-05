package main

import "testing"

var mockApp = &App{
	services: map[string]string{
		"downloader": "http://localhost:4002/pic.sum/random/download",
	},
}

func Test(t *testing.T) {
	err := mockApp.downloadRandomNRequest("127.0.0.1", 100)
	if err != nil {
		t.Fatal(err)
	}
}
