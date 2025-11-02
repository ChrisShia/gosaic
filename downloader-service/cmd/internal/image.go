package internal

import "image"

//func AverageColor(img image.Image) ([3]float64, error) {
//	bounds := img.Bounds()
//	r, g, b := 0.0, 0.0, 0.0
//
//	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
//		for x := bounds.Min.X; x < bounds.Max.X; x++ {
//			r1, g1, b1, _ := img.At(x, y).RGBA()
//			r, g, b = r+float64(r1), g+float64(g1), b+float64(b1)
//		}
//	}
//
//	totalPixels := float64(bounds.Max.X * bounds.Max.Y)
//	return [3]float64{r / totalPixels, g / totalPixels, b / totalPixels}, nil
//}

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
