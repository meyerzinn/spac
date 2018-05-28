package stars

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"math"
	"sync"
)

const (
	Seed = 0x9d2c5681
)

var (
	imdMu = new(sync.Mutex)
	imd   = imdraw.New(nil)
)

func Static(win *pixelgl.Window) {
	imdMu.Lock()
	defer imdMu.Unlock()
	imd.Clear()
	imd.Color = colornames.Darkgray
	Draw(imd, pixel.ZV, win.Bounds(), 4)
	imd.Color = colornames.Gray
	Draw(imd, pixel.ZV, win.Bounds(), 2)
	imd.Color = colornames.White
	Draw(imd, pixel.ZV, win.Bounds(), 1)
	imd.Draw(win)
}

func Draw(imd *imdraw.IMDraw, cam pixel.Vec, bounds pixel.Rect, starscale int) {
	w := int(bounds.W())
	h := int(bounds.H())
	size := int(math.Max(bounds.W(), bounds.H())) / starscale
	xoff := int(cam.X) - w/2
	yoff := int(cam.Y) - h/2
	sx := xoff/size*size - size
	sy := yoff/size*size - size
	for i := sx; i <= xoff+w+size; i += size {
		for j := sy; j <= yoff+h+size; j += size {
			hash := mix(Seed, i, j)
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
