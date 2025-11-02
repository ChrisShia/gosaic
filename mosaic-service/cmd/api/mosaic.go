package main

import (
	"errors"
	"image"
	"image/draw"
	"sync"

	"github.com/ChrisShia/mosaic/cmd/internal"
)

var (
	ErrInvalidTilesRepository = errors.New("invalid tiles repository")
)

type builder struct {
	tiles          internal.TileRepository
	originalImg    image.Image
	tileWidth      int
	tileWidthFloat float64
	mosaicImg      draw.Image
}

func NewMosaicBuilder(tiles internal.TileRepository, originalImg image.Image, tileWidth int) *builder {
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

	return b.mosaicImg, nil
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

func (b *builder) putTileAt(sp point, dst drawer) (*rect, error) {
	r := rect{Min: sp, Max: sp.Add(point{X: b.tileWidth, Y: b.tileWidth})}

	imageFromRepository, err := b.findImageByAverageColor(r)
	if err != nil {
		return nil, err
	}

	resizedImg, err := resize(b.tileWidthFloat, imageFromRepository)
	if err != nil {
		return nil, err
	}

	paintedRectangle := b.drawTileAtXY(resizedImg, sp, dst)

	return &paintedRectangle, nil
}

func (b *builder) drawTileAtXY(src image.Image, sp image.Point, dst drawer) image.Rectangle {
	tileBounds := src.Bounds()
	tileBoundsInOriginalFrame := rect{Min: sp, Max: sp.Add(tileBounds.Max)}

	b.drawTile(src, tileBoundsInOriginalFrame, dst)

	return tileBoundsInOriginalFrame
}

func (b *builder) findImageByAverageColor(r rect) (image.Image, error) {
	color := internal.AverageRGBArea(b.originalImg, r.Min.X, r.Max.X, r.Min.Y, r.Max.Y)

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

func resize(newWidth float64, img image.Image) (image.Image, error) {
	//bounds := img.Bounds()
	//scaleFactor := newWidth / float64(bounds.Dx())
	t := resizeByAveragePooling(img, int(newWidth))
	//t, err := ResizeGoCV(img, scaleFactor, gocv.InterpolationArea)
	//if err != nil {
	//	return nil, err
	//}

	return t, nil
}
