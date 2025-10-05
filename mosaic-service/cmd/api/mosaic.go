package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
	"os"
	"sync"
)

//TODO: tileSize should be tested against the original img size as
//to whether it is an exact divisor of the img width...

type Image struct {
	bs           []byte
	averageColor [3]float64
}

func (app *App) mosaic(originalImgBytes io.Reader, reqIP string, tileSize int) {
	original, _, _ := image.Decode(originalImgBytes)
	//bounds := original.Bounds()

	//TODO: This should be the source of tiles...a temporary and short-lived
	//		may be appropriate.
	//db := cloneTilesDB()

	//c := createWorkersAndCombine(original, tileSource, tileSize, bounds)

	buf1 := new(bytes.Buffer)
	jpeg.Encode(buf1, original, nil)
	//originalStr := base64.StdEncoding.EncodeToString(buf1.Bytes())

}

func createWorkersAndCombine(original image.Image, reqIP string, tileSize int, bounds image.Rectangle) {
	c1 := cut(original, reqIP, tileSize, bounds.Min.X, bounds.Min.Y, bounds.Max.X/2, bounds.Max.Y/2)
	c2 := cut(original, reqIP, tileSize, bounds.Max.X/2, bounds.Min.Y, bounds.Max.X, bounds.Max.Y/2)
	c3 := cut(original, reqIP, tileSize, bounds.Min.X, bounds.Max.Y/2, bounds.Max.X/2, bounds.Max.Y)
	c4 := cut(original, reqIP, tileSize, bounds.Max.X/2, bounds.Max.Y/2, bounds.Max.X, bounds.Max.Y)
	combine(bounds, c1, c2, c3, c4)
}

func combine(r image.Rectangle, c1, c2, c3, c4 <-chan image.Image) {

	go func() {
		var wg sync.WaitGroup
		img := image.NewNRGBA(r)
		cpy := func(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point) {
			draw.Draw(dst, r, src, sp, draw.Src)
			wg.Done()
		}

		wg.Add(4)
		var s1, s2, s3, s4 image.Image
		var ok1, ok2, ok3, ok4 bool
		for {
			select {
			case s1, ok1 = <-c1:
				go cpy(img, s1.Bounds(), s1, image.Point{X: r.Min.X, Y: r.Min.Y})
				//TODO: I put this nil assignment so that channels are closed without damaging the select
				//procedure...(closed channels are always ready to be read from and so are always selected)
				c1 = nil
			case s2, ok2 = <-c2:
				go cpy(img, s2.Bounds(), s2, image.Point{X: r.Max.X / 2, Y: r.Min.Y})
				c2 = nil
			case s3, ok3 = <-c3:
				go cpy(img, s3.Bounds(), s3, image.Point{X: r.Min.X, Y: r.Max.Y / 2})
				c3 = nil
			case s4, ok4 = <-c4:
				go cpy(img, s4.Bounds(), s4, image.Point{X: r.Max.X / 2, Y: r.Max.Y / 2})
				c4 = nil
			}
			if ok1 && ok2 && ok3 && ok4 {
				break
			}
		}

		buf := new(bytes.Buffer)

		jpeg.Encode(buf, img, nil)

		//toString := base64.StdEncoding.EncodeToString(buf.Bytes())
	}()
}

func cut(original image.Image, reqIP string, tileSize, minX, minY, maxX, maxY int) <-chan image.Image {
	c := make(chan image.Image)
	sp := image.Point{}

	go func() {
		defer close(c)
		sectorImg := image.NewNRGBA(image.Rect(minX, minY, maxX, maxY))
		for y := minY; y < maxY; y = y + tileSize {
			for x := minX; x < maxX; x = x + tileSize {
				//color := RGBAt(original, x, y)
				//nearest := db.nearest(color)
				//file, err := os.Open(nearest)

				file, err := os.Open("")
				if err == nil {
					img, _, err := image.Decode(file)
					if err == nil {
						t := resizeByNearestNeighbour(img, tileSize)
						tile := t.SubImage(t.Bounds())
						tileBounds := image.Rect(x, y, x+tileSize, y+tileSize)
						draw.Draw(sectorImg, tileBounds, tile, sp, draw.Src)
					} else {
						fmt.Println("error:", err)
					}
				}
			}
		}
	}()

	return c
}

func RGBAt(img image.Image, x int, y int) [3]float64 {
	r, g, b, _ := img.At(x, y).RGBA()
	color := [3]float64{float64(r), float64(g), float64(b)}
	return color
}

//func (app *App) imageRetriever(reqIP string, averageColor [3]float64) image.Image {
//	app.cfg.Redis.Client.HGet
//}
