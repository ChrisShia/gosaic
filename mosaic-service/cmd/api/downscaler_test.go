package main

import (
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"os"
	"testing"
)

const testImage200300 = "test/test_image_200300.jpg"

//NOTE: the resized images are embedded on a frame with larger dimensions.

func Test_resize(t *testing.T) {
	t.Run("", func(t *testing.T) {
		frame := image.NewNRGBA(image.Rect(0, 0, 100, 100))
		img_200_300 := testImage()
		newWidth := 20
		imgResized := resizeByNearestNeighbour(img_200_300, newWidth)
		tile := imgResized.SubImage(imgResized.Bounds())
		offSetX := 20
		offSetY := 20
		tileBoundsInFrame := image.Rect(offSetX, offSetY, offSetX+newWidth, offSetY+30)
		draw.Draw(frame, tileBoundsInFrame, tile, image.Point{}, draw.Src)
		dstFile, err := os.Create("./test/resized_nearest_neighbour.jpg")
		if err != nil {
			log.Fatal(err)
		}
		jpeg.Encode(dstFile, frame, nil)
	})
}

func Test_ResizeByAveragePooling(t *testing.T) {
	t.Run("", func(t *testing.T) {
		frame := image.NewNRGBA(image.Rect(0, 0, 100, 100))
		img_200_300 := testImage()
		newWidth := 20
		imgResized := resizeByAveragePooling(img_200_300, newWidth)
		tile := imgResized.SubImage(imgResized.Bounds())
		offSetX := 20
		offSetY := 20
		tileBoundsInFrame := image.Rect(offSetX, offSetY, offSetX+newWidth, offSetY+30)
		draw.Draw(frame, tileBoundsInFrame, tile, image.Point{}, draw.Src)
		dstFile, err := os.Create("./test/resized_average_pooling.jpg")
		if err != nil {
			log.Fatal(err)
		}
		jpeg.Encode(dstFile, frame, nil)
	})
}

func testImage() image.Image {
	file, err := os.Open(testImage200300)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	return img
}
