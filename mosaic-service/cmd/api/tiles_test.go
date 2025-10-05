package main

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"path/filepath"
)

const tilesTestSrcDir = "Downloads"
const originImg = "test/original/img.jpeg"

func imgStream(in <-chan []byte) <-chan image.Image {
	out := make(chan image.Image)

	go func() {
		defer close(out)
		for v := range in {
			reader := bytes.NewReader(v)
			img, _, err := image.Decode(reader)
			if err != nil {
				fmt.Println(err)
				continue
			}
			out <- img
		}
	}()

	return out
}

func srcDirContentByteStream() <-chan []byte {
	out := make(chan []byte)
	dirEntries, err := os.ReadDir(tilesTestSrcDir)
	if err != nil {
		panic(err)
	}

	go func() {
		defer close(out)
		for _, dirEntry := range dirEntries {
			file, err := os.ReadFile(filepath.Join(tilesTestSrcDir, dirEntry.Name()))
			if err != nil {
				fmt.Println(err)
				continue
			}

			out <- file
		}
	}()

	return out
}
