package main

import (
	"errors"
	"image"
	"image/draw"
	"sync"

	"gocv.io/x/gocv"
)

type TileRepository interface {
	Image(ac [3]float64) (image.Image, error)
}

type builder struct {
	tiles          TileRepository
	originalImg    image.Image
	tileWidth      int
	tileWidthFloat float64
	mosaicImg      draw.Image
}

var (
	ErrInvalidTilesRepository = errors.New("invalid tiles repository")
)

func NewMosaicBuilder(tiles TileRepository, originalImg image.Image, tileWidth int) *builder {

	return &builder{
		tiles:          tiles,
		originalImg:    originalImg,
		tileWidth:      tileWidth,
		tileWidthFloat: float64(tileWidth),
		mosaicImg:      image.NewNRGBA(originalImg.Bounds()),
	}
}

func (b *builder) Mosaic() (image.Image, error) {
	if b.tiles == nil {
		return nil, ErrInvalidTilesRepository
	}

	b.mosaic()

	return nil, nil
}

func (b *builder) mosaic() {
	bounds := b.originalImg.Bounds()

	//TODO: abstract away...call sectorWorker in a loop over a slice of bounds
	//TODO: examine processor number and divide bounds accordingly
	//TODO use recover
	c1 := b.sectorWorker(bounds.Min.X, bounds.Min.Y, bounds.Max.X/2, bounds.Max.Y/2)
	c2 := b.sectorWorker(bounds.Max.X/2, bounds.Min.Y, bounds.Max.X, bounds.Max.Y/2)
	c3 := b.sectorWorker(bounds.Min.X, bounds.Max.Y/2, bounds.Max.X/2, bounds.Max.Y)
	c4 := b.sectorWorker(bounds.Max.X/2, bounds.Max.Y/2, bounds.Max.X, bounds.Max.Y)
	b.combineSingleReceiveChannels(c1, c2, c3, c4)
}

func (b *builder) combineSingleReceiveChannels(c1, c2, c3, c4 <-chan image.Image) {
	r := b.mosaicImg.Bounds()

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		cpy := func(dst drawer, r image.Rectangle, src image.Image, sp image.Point) {
			draw.Draw(dst, r, src, sp, draw.Src)
			wg.Done()
		}

		var s1, s2, s3, s4 image.Image
		var ok1, ok2, ok3, ok4 bool
		for {
			select {
			case s1, ok1 = <-c1:
				go cpy(b.mosaicImg, s1.Bounds(), s1, image.Point{X: r.Min.X, Y: r.Min.Y})
				c1 = nil
			case s2, ok2 = <-c2:
				go cpy(b.mosaicImg, s2.Bounds(), s2, image.Point{X: r.Max.X / 2, Y: r.Min.Y})
				c2 = nil
			case s3, ok3 = <-c3:
				go cpy(b.mosaicImg, s3.Bounds(), s3, image.Point{X: r.Min.X, Y: r.Max.Y / 2})
				c3 = nil
			case s4, ok4 = <-c4:
				go cpy(b.mosaicImg, s4.Bounds(), s4, image.Point{X: r.Max.X / 2, Y: r.Max.Y / 2})
				c4 = nil
			}
			if ok1 && ok2 && ok3 && ok4 {
				break
			}
		}
	}()

	wg.Wait()
}

func (b *builder) sectorWorker(minX, minY, maxX, maxY int) <-chan image.Image {
	c := make(chan image.Image)

	go func() {
		defer close(c)

		secRect := image.Rect(minX, minY, maxX, maxY)
		sectorImg := image.NewNRGBA(secRect)
		b.fillWithTiles(sectorImg)
		c <- sectorImg
	}()

	return c
}

func (b *builder) fillWithTiles(dst drawer) {
	bounds := dst.Bounds()
	var wg sync.WaitGroup

	for x := bounds.Min.X; x < bounds.Max.X; x = x + b.tileWidth {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := bounds.Min.Y; y < bounds.Max.Y; {
				sp := point{X: x, Y: y}

				r, err := b.putTileAt(sp, dst)
				if err != nil {
					continue
				}

				y = r.Max.Y
			}
		}()
	}

	wg.Wait()
}

type drawer interface {
	draw.Image
}

func (b *builder) putTileAt(sp point, dst drawer) (rect, error) {
	r := rect{Min: sp, Max: sp.Add(point{X: b.tileWidth, Y: b.tileWidth})}

	imageFromRepository, err := b.findImageByAverageColor(r)

	resizedImg, err := resize(b.tileWidthFloat, imageFromRepository)

	paintedRectangle := b.drawTileAtXY(resizedImg, sp, dst)

	return paintedRectangle, err
}

func (b *builder) drawTileAtXY(resizedImg image.Image, sp image.Point, sectorImg drawer) image.Rectangle {
	tileBounds := resizedImg.Bounds()
	tileBoundsInOriginalFrame := rect{Min: sp, Max: sp.Add(tileBounds.Max)}

	b.drawTile(resizedImg, tileBoundsInOriginalFrame, sectorImg)

	return tileBoundsInOriginalFrame
}

func (b *builder) findImageByAverageColor(r rect) (image.Image, error) {
	color := AverageRGBAt(b.originalImg, r.Min.X, r.Max.X, r.Min.Y, r.Max.Y)

	imgFromTileRepository, err := b.tiles.Image(color)
	if err != nil {
		return nil, err
	}

	return imgFromTileRepository, err
}

type point = image.Point
type rect = image.Rectangle

func (b *builder) drawTile(tileImg image.Image, r image.Rectangle, sectorImg drawer) {
	zeroPoint := point{}
	draw.Draw(sectorImg, r, tileImg, zeroPoint, draw.Src)
}

func resize(newWidth float64, img image.Image) (*image.RGBA, error) {
	bounds := img.Bounds()
	scaleFactor := newWidth / float64(bounds.Dx())
	t, err := ResizeGoCV(img, scaleFactor, gocv.InterpolationArea)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func AverageRGBAt(img image.Image, xMin, xMax, yMin, yMax int) [3]float64 {
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
