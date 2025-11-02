package internal

import "image"

type TileRepository interface {
	Image(ac [3]float64) (image.Image, error)
}
