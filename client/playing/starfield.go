package playing

import (
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel"
	"math"
)

const (
	StarSeed     = 0x9d2c5680
	StarTileSize = 256
)

func drawStars(imd *imdraw.IMDraw, off pixel.Vec, bounds pixel.Rect, starscale int) {
	//imd.Precision = 30
	size := int(StarTileSize / starscale)
	offx := int(math.Floor(off.X))
	offy := int(math.Floor(off.Y))
	w := int(bounds.W())
	h := int(bounds.H())
	sx := ((offx-w/2)/size)*size - size
	sy := ((offy-h/2)/size)*size - size
	for i := sx; i <= w+sx+size*3; i += size {
		for j := sy; j <= h+sy+size*3; j += size {
			hash := mix(StarSeed, i, j)
			for n := 0; n < 3; n++ {
				px := (hash % size) + (i - offx)
				hash >>= 3
				py := (hash % int(size)) + (j - offy)
				hash >>= 3
				imd.Push(pixel.V(float64(px), float64(py)), )
				imd.Circle(1, 1)
			}
		}
	}
}
func mix(a, b, c int) int {
	a = a - b
	a = a - c
	a = a ^ (c >> 13)
	b = b - c
	b = b - a
	b = b ^ (a << 8)
	c = c - a
	c = c - b
	c = c ^ (b >> 13)
	a = a - b
	a = a - c
	a = a ^ (c >> 12)
	b = b - c
	b = b - a
	b = b ^ (a << 16)
	c = c - a
	c = c - b
	c = c ^ (b >> 5)
	a = a - b
	a = a - c
	a = a ^ (c >> 3)
	b = b - c
	b = b - a
	b = b ^ (a << 10)
	c = c - a
	c = c - b
	c = c ^ (b >> 15)
	return c
}
