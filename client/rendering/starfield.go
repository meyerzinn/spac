package rendering

import (
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel"
	"math"
)

const (
	StarSeed     int = 0x9d2c5681
	StarTileSize     = 512
)

func drawStars(imd *imdraw.IMDraw, cam pixel.Vec, bounds pixel.Rect, starscale int) {
	size := StarTileSize / starscale
	xoff := int(math.Floor(cam.X))
	yoff := int(math.Floor(cam.Y))
	w := int(bounds.W())
	h := int(bounds.H())
	sx := ((xoff-w/2)/size)*size - size
	sy := ((yoff-h/2)/size)*size - size
	for i := sx; i <= sx+w+size*3; i += size {
		for j := sy; j <= sy+h+size*3; j += size {
			hash := mix(StarSeed, i, j)
			for n := 0; n < 3; n++ {
				px := (hash % size) + i
				hash >>= 3
				py := (hash % size) + j
				hash >>= 3
				imd.Push(pixel.Vec{float64(px), float64(py)})
				imd.Circle(1, 0)
			}
		}
	}
}

func mix(a, b, c int) int {
	a -= b
	a -= c
	a ^= c >> 13
	b -= c
	b -= a
	b ^= a << 8
	c -= a
	c -= b
	c ^= b >> 13
	a -= b
	a -= c
	a ^= c >> 12
	b -= c
	b -= a
	b ^= a << 16
	c -= a
	c -= b
	c ^= b >> 5
	a -= b
	a -= c
	a ^= c >> 3
	b -= c
	b -= a
	b ^= a << 10
	c -= a
	c -= b
	c ^= b >> 15
	return c
}
