package main

import (
	"image"
	"image/color"
)

func resizeByNearestNeighbour(inputImg image.Image, newWidth int) image.NRGBA {
	bounds := inputImg.Bounds()
	scalingFactor := bounds.Dx() / newWidth
	x0 := bounds.Min.X / scalingFactor
	y0 := bounds.Min.X / scalingFactor
	x1 := bounds.Max.X / scalingFactor
	y1 := bounds.Max.Y / scalingFactor
	out := image.NewNRGBA(image.Rect(x0, y0, x1, y1))

	for j, y := bounds.Min.Y, bounds.Min.Y; y < bounds.Max.Y; j, y = j+1, y+scalingFactor {
		for i, x := bounds.Min.X, bounds.Min.X; x < bounds.Max.X; i, x = i+1, x+scalingFactor {
			r, g, b, a := inputImg.At(x, y).RGBA()
			out.SetNRGBA(i, j, color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)})
		}
	}

	return *out
}

func resizeByAveragePooling(inputImg image.Image, newWidth int) image.NRGBA {
	bounds := inputImg.Bounds()
	scaleFactor := bounds.Dx() / newWidth
	x0 := bounds.Min.X / scaleFactor
	y0 := bounds.Min.X / scaleFactor
	x1 := bounds.Max.X / scaleFactor
	y1 := bounds.Max.Y / scaleFactor
	out := image.NewNRGBA(image.Rect(x0, y0, x1, y1))
	for j, y := 0, bounds.Min.Y; y < bounds.Max.Y; j, y = j+1, y+scaleFactor {
		for i, x := 0, bounds.Min.X; x < bounds.Max.X; i, x = i+1, x+scaleFactor {

			var rSum, gSum, bSum, aSum uint64
			var count uint64

			for yy := y; yy < y+scaleFactor && yy < bounds.Max.Y; yy++ {
				for xx := x; xx < x+scaleFactor && xx < bounds.Max.X; xx++ {
					r, g, b, a := inputImg.At(xx, yy).RGBA()
					rSum += uint64(r >> 8)
					gSum += uint64(g >> 8)
					bSum += uint64(b >> 8)
					aSum += uint64(a >> 8)
					count++
				}
			}

			if count == 0 {
				continue
			}

			out.SetNRGBA(i, j, color.NRGBA{
				R: uint8(rSum / count),
				G: uint8(gSum / count),
				B: uint8(bSum / count),
				A: uint8(aSum / count),
			})
		}
	}

	return *out
}
