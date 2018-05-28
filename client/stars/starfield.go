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
	gimdMu = new(sync.Mutex)
	gImd   = imdraw.New(nil)
)

func Static(win *pixelgl.Window) {
	gimdMu.Lock()
	defer gimdMu.Unlock()
	gImd.Clear()
	gImd.Color = colornames.Darkgray
	Draw(gImd, pixel.ZV, win.Bounds(), 4)
	gImd.Color = colornames.Gray
	Draw(gImd, pixel.ZV, win.Bounds(), 2)
	gImd.Color = colornames.White
	Draw(gImd, pixel.ZV, win.Bounds(), 1)
	gImd.Draw(win)
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
