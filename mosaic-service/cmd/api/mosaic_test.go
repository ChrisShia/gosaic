package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func Test_Mosaic(t *testing.T) {
	decode := func(file *os.File) (image.Image, error) {
		return png.Decode(file)
	}

	img := Image("../../test_image_700.png", decode)

	mosaic, err := NewMosaicBuilder(mockTileRepository_, img, 20).Mosaic()
	if err != nil {
		t.Fatal(err)
		return
	}

	resultFile, err := os.Create("mosaic.png")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer resultFile.Close()

	err = png.Encode(resultFile, mosaic)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func Test_combineSectorImages(t *testing.T) {
	bounds := image.Rect(0, 0, 2000, 2000)

	nrgbaImg := image.NewNRGBA(bounds)

	gi := &gridImage{img: nrgbaImg}
	b := &builder{tiles: mockInfiniteTileRepository_, originalImg: gi, tileWidth: 20, tileWidthFloat: float64(20), mosaicImg: gi}

	c1 := b.sectorWorker(bounds.Min.X, bounds.Min.Y, bounds.Max.X/2, bounds.Max.Y/2)
	c2 := b.sectorWorker(bounds.Max.X/2, bounds.Min.Y, bounds.Max.X, bounds.Max.Y/2)
	c3 := b.sectorWorker(bounds.Min.X, bounds.Max.Y/2, bounds.Max.X/2, bounds.Max.Y)
	c4 := b.sectorWorker(bounds.Max.X/2, bounds.Max.Y/2, bounds.Max.X, bounds.Max.Y)

	b.combineSingleReceiveChannels(c1, c2, c3, c4)

	saveInTestDir("testCombineSingleReceiveChannels", b.mosaicImg)
}

func Test_SectorWorker(t *testing.T) {
	bounds := image.Rect(0, 0, 2000, 2000)

	nrgbaImg := image.NewNRGBA(bounds)

	gi := &gridImage{img: nrgbaImg}
	b := &builder{tiles: mockTileRepository_, originalImg: gi, tileWidth: 40, tileWidthFloat: float64(40)}

	imageChan := b.sectorWorker(bounds.Min.X, bounds.Min.Y, bounds.Max.X/2, bounds.Max.Y/2)
	image := <-imageChan
	saveInTestDir("testSectorWorker", image)
}

func Test_FillWithTiles(t *testing.T) {
	bounds := image.Rect(0, 0, 400, 400)
	nrgbaImg := image.NewNRGBA(bounds)

	gi := &gridImage{img: nrgbaImg}
	b := &builder{tiles: mockTileRepository_, originalImg: gi, tileWidth: 40, tileWidthFloat: float64(40)}

	b.fillWithTiles(gi)

	gi.Grid(40)
	saveInTestDir("testFillWithTiles", gi)
}

func Test_drawTileAtXY(t *testing.T) {
	bounds := image.Rect(0, 0, 400, 400)
	nrgbaImg := image.NewNRGBA(bounds)

	gi := &gridImage{img: nrgbaImg}
	b := &builder{tiles: mockTileRepository_, originalImg: gi}

	tileImage, _ := b.tiles.Image([3]float64{0, 0, 0})

	sp := point{X: 100, Y: 100}
	paintedRect := b.drawTileAtXY(tileImage, sp, gi)
	if paintedRect.Min.X != sp.X || paintedRect.Min.Y != sp.Y {
		t.Errorf("Expected paintedRect.Min=(100,100) got %v", paintedRect.Min)
	}

	sp = point{X: 200, Y: 0}
	paintedRect = b.drawTileAtXY(tileImage, sp, gi)
	if paintedRect.Min.X != sp.X || paintedRect.Min.Y != sp.Y {
		t.Errorf("Expected paintedRect.Min=(200,0) got %v", paintedRect.Min)
	}
	if paintedRect.Max.X != sp.X+tileImage.Bounds().Dx() || paintedRect.Max.Y != sp.Y+tileImage.Bounds().Dy() {
		t.Errorf("Expected paintedRect.Min=(200,0) got %v", paintedRect.Min)
	}

	gi.Grid(20)

	saveInTestDir("drawTileAtXY.png", gi)
}

func Test_drawTile(t *testing.T) {
	testName := "drawTile"
	bounds := image.Rect(0, 0, 400, 400)
	nrgbaImg := image.NewNRGBA(bounds)

	result := &gridImage{img: nrgbaImg}
	b := &builder{tiles: mockTileRepository_, originalImg: result}

	tileImage, _ := b.tiles.Image([3]float64{0, 0, 0})

	b.drawTile(tileImage, bounds, result)

	result.Grid(20)

	saveInTestDir(testName, nrgbaImg)
}

func saveInTestDir(testName string, nrgbaImg image.Image) {
	resultFile, err := os.Create(fmt.Sprintf("../../test/%s.png", testName))
	if err != nil {
		log.Fatal(err)
	}
	defer resultFile.Close()
	err = png.Encode(resultFile, nrgbaImg)
	if err != nil {
		log.Fatal(err)
	}
}

type MockImage struct {
	bounds     image.Rectangle
	color      [][]color.Color
	colorModel color.Model
}

func (mi *MockImage) ColorModel() color.Model {
	return mi.colorModel
}

func (mi *MockImage) Bounds() image.Rectangle {
	return mi.bounds
}

func (mi *MockImage) At(x, y int) color.Color {
	return mi.color[x][y]
}

func (mi *MockImage) Set(x, y int, c color.Color) {
	mi.color[x][y] = c
}

func Test_grid(t *testing.T) {
	grid := gridImage{img: image.NewNRGBA(image.Rect(0, 0, 401, 401))}
	grid.Grid(20)
	file, _ := os.Create("../../test/grid.png")
	defer file.Close()
	err := png.Encode(file, grid.img)
	if err != nil {
		t.Fatal(err)
	}
}

type gridImage struct {
	img *image.NRGBA
}

func (gi *gridImage) ColorModel() color.Model {
	return gi.img.ColorModel()
}

func (gi *gridImage) Bounds() image.Rectangle {
	return gi.img.Bounds()
}

func (gi *gridImage) At(x, y int) color.Color {
	return gi.img.At(x, y)
}

func (gi *gridImage) Set(x, y int, c color.Color) {
	gi.img.Set(x, y, c)
}

func (gi *gridImage) Grid(tileWidth int) {
	for x := 0; x < gi.img.Bounds().Dx(); x = x + tileWidth {
		for y := 0; y < gi.img.Bounds().Dy(); y++ {
			gi.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	for y := 0; y <= gi.img.Bounds().Dy(); y = y + tileWidth {
		for x := 0; x < gi.img.Bounds().Dx(); x++ {
			gi.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
}

type MockInfiniteTileRepository struct {
	images []image.Image
	len    int
}

func (mi *MockInfiniteTileRepository) Image(ac [3]float64) (image.Image, error) {
	return mi.Random(), nil
}

func (mi *MockInfiniteTileRepository) Random() image.Image {
	randomIndex := rand.Intn(mi.len)

	return mi.images[randomIndex]
}

type MockTileRepository struct {
	images []image.Image
}

func (m *MockTileRepository) Pop() image.Image {
	popped := m.images[0]
	m.images = m.images[1:]
	return popped
}

func (m *MockTileRepository) Image(ac [3]float64) (image.Image, error) {
	return m.Pop(), nil
}

var mockTileRepository_ = NewMockTileRepository()
var mockInfiniteTileRepository_ = NewMockInfiniteTileRepository()

func NewMockTileRepository() TileRepository {
	return &MockTileRepository{images: images()}
}

func NewMockInfiniteTileRepository() TileRepository {
	list := images()
	return &MockInfiniteTileRepository{images: list, len: len(list)}
}

func images() []image.Image {
	list := make([]image.Image, 0)

	dirEntries, err := os.ReadDir("../../../Downloads")
	if err != nil {
		panic(err)
	}

	decode := func(file *os.File) (image.Image, error) {
		return jpeg.Decode(file)
	}

	for _, dirEntry := range dirEntries {
		img := Image(filepath.Join("../../../Downloads", dirEntry.Name()), decode)
		list = append(list, img)
	}
	return list
}

func imageAverageRGB(img image.Image) [3]float64 {
	bounds := img.Bounds()
	return AverageRGBAt(img, bounds.Min.X, bounds.Max.X, bounds.Min.Y, bounds.Max.Y)
}

func Image(path string, decode func(file *os.File) (image.Image, error)) image.Image {
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

func ImageBufferForTestImage(path string, decoder func(file *os.File) (image.Image, error)) (*bytes.Buffer, error) {
	originalImg := Image(path, decoder)
	bs := make([]byte, 0)
	imgBuf := bytes.NewBuffer(bs)

	err := png.Encode(imgBuf, originalImg)
	if err != nil {
		return nil, err
	}

	return imgBuf, nil
}
