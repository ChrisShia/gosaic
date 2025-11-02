package main

import (
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"testing"
)

//NOTE: the resized images are embedded on a frame with larger dimensions.

func Test_ResizeGoCV(t *testing.T) {
	//var tt = []struct {
	//	name   string
	//	method gocv.InterpolationFlags
	//}{
	//	{"NearestNeighbor", gocv.InterpolationNearestNeighbor},
	//	{"Linear", gocv.InterpolationLinear},
	//	{"Cubic", gocv.InterpolationCubic},
	//	{"Area", gocv.InterpolationArea},
	//	{"Lanczos4", gocv.InterpolationLanczos4},
	//	//{"Max", gocv.InterpolationMax},
	//}
	//
	//for _, tc := range tt {
	//	t.Run(tc.name, func(t *testing.T) {
	//		img_700 := pngImage("../../test/test_image_700.png")
	//		resized, err := ResizeGoCV(img_700, 0.05, tc.method)
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//
	//		newImageFileName := fmt.Sprintf("../../test/resized_go_cv_%s.jpg", tc.name)
	//		dstFile, err := os.Create(newImageFileName)
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//		err = png.Encode(dstFile, resized)
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//	})
	//}
}

func Test_resize(t *testing.T) {
	t.Run("", func(t *testing.T) {
		frame := image.NewNRGBA(image.Rect(0, 0, 100, 100))
		img_200_300 := jpegImage("../../test/test_image_200_300.jpg")
		newWidth := 20
		imgResized := resizeByNearestNeighbour(img_200_300, newWidth)
		tile := imgResized.SubImage(imgResized.Bounds())
		offSetX := 20
		offSetY := 20
		tileBoundsInFrame := image.Rect(offSetX, offSetY, offSetX+newWidth, offSetY+30)
		draw.Draw(frame, tileBoundsInFrame, tile, image.Point{}, draw.Src)
		dstFile, err := os.Create("../../test/resized_nearest_neighbour.jpg")
		if err != nil {
			log.Fatal(err)
		}
		jpeg.Encode(dstFile, frame, nil)
	})
}

// TODO: rewrite test
func Test_ResizeByAveragePooling(t *testing.T) {
	t.Run("", func(t *testing.T) {
		//frame := image.NewNRGBA(image.Rect(0, 0, 100, 100))
		//img_200_300 := jpegImage("../../test/test_image_200_300.jpg")
		//newWidth := 20
		//imgResized := resizeByAveragePooling(img_200_300, newWidth)
		//tile := imgResized.SubImage(imgResized.Bounds())
		//offSetX := 20
		//offSetY := 20
		//tileBoundsInFrame := image.Rect(offSetX, offSetY, offSetX+newWidth, offSetY+30)
		//draw.Draw(frame, tileBoundsInFrame, tile, image.Point{}, draw.Src)
		//dstFile, err := os.Create("../../test/resized_average_pooling.jpg")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//jpeg.Encode(dstFile, frame, nil)
	})
}

func pngImage(path string) image.Image {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	return img
}

func jpegImage(path string) image.Image {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	return img
}
