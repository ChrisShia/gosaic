package internal

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"io"
	"log"
	"os"
)

func Base64StringToImage(str string) (image.Image, error) {
	decodedLen := base64.StdEncoding.DecodedLen(len(str))
	p := make([]byte, decodedLen)
	_, err := base64.StdEncoding.Decode(p, []byte(str))
	if err != nil {
		return nil, err
	}

	return PngDecode(bytes.NewBuffer(p))
}

func PngDecode(r io.Reader) (image.Image, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func ImageToBase64String(img image.Image) (string, error) {
	jpegEncoder := func(w io.Writer, img image.Image) error {
		return png.Encode(w, img)
	}

	imgBuf, err := imageToBytes(img, jpegEncoder)
	if err != nil {
		return "", err
	}

	base64Str := base64.StdEncoding.EncodeToString(imgBuf.Bytes())

	return base64Str, nil
}

func imageToBytes(img image.Image, encoder func(io.Writer, image.Image) error) (*bytes.Buffer, error) {
	imgBuf := new(bytes.Buffer)

	err := encoder(imgBuf, img)
	if err != nil {
		return nil, err
	}
	return imgBuf, nil
}

func ImageAverageRGB(img image.Image) [3]float64 {
	bounds := img.Bounds()
	return AverageRGBArea(img, bounds.Min.X, bounds.Max.X, bounds.Min.Y, bounds.Max.Y)
}

func AverageRGBArea(img image.Image, xMin, xMax, yMin, yMax int) [3]float64 {
	var rSum, gSum, bSum uint64
	var count uint64

	for yy := yMin; yy < yMax; yy++ {
		for xx := xMin; xx < xMax; xx++ {
			r, g, b, _ := img.At(xx, yy).RGBA()
			rSum += uint64(r >> 8)
			gSum += uint64(g >> 8)
			bSum += uint64(b >> 8)
			count++
		}
	}

	if count == 0 {
		return [3]float64{0, 0, 0}
	}

	rAve := float64(rSum) / float64(count)
	gAve := float64(gSum) / float64(count)
	bAve := float64(bSum) / float64(count)

	return [3]float64{rAve, gAve, bAve}
}

func RGBAt(img image.Image, x int, y int) [3]float64 {
	r, g, b, _ := img.At(x, y).RGBA()
	color := [3]float64{float64(r), float64(g), float64(b)}
	return color
}

func ImageDecodeFunc(path string, decode func(file *os.File) (image.Image, error)) image.Image {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, err := decode(file)
	if err != nil {
		log.Fatal(err)
	}
	return img
}
