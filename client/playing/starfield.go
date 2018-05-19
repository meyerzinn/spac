package playing

import (
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel"
	"math"
)

const (
	StarSeed     int = 0x9d2c5681
	StarTileSize     = 256
)

func drawStars(imd *imdraw.IMDraw, cam pixel.Vec, bounds pixel.Rect, starscale int) {
	size := StarTileSize / starscale
	xoff := int(math.Floor(cam.X))
	yoff := int(math.Floor(cam.Y))
	w := int(bounds.W())
	h := int(bounds.H())
	sx := ((xoff-w/2)/size)*size - size
	sy := ((yoff-h/2)/size)*size - size

	//fmt.Println(step, yoff, sy, sy+h+size*3, cam.Y+bounds.H()/2, yoff-sy)
	for i := sx; i <= sx+w+size*3; i += size {
		for j := sy; j <= sy+h+size*3; j += size {
			hash := mix(StarSeed, i, j)
			//fmt.Println(i, j, hash%size)
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
	//for x := s.X; x <= s.X+bounds.W(); x += float64(size) {
	//	for y := s.Y; y <= s.Y+bounds.H(); y += float64(size) {
	//		hash := mix(StarSeed, int(math.Floor(x)), int(math.Floor(y)))
	//		for n := 0; n < 3; n++ {
	//			px := float64(hash%size) + x + s.X
	//			hash >>= 3
	//			py := float64(hash%size) + y + s.Y
	//			hash >>= 3
	//			imd.Push(pixel.Vec{px, py})
	//			imd.Circle(1, 0)
	//		}
	//	}
	//}
}

//
//func rot(x int, k uint) int {
//	return ((x) << (k)) | ((x) >> (32 - (k)))
//}
//
//func mix(a, b, c int) int {
//	a -= c
//	a ^= rot(c, 4)
//	c += b
//	b -= a
//	b ^= rot(a, 6)
//	a += c
//	c -= b
//	c ^= rot(b, 8)
//	b += a
//	a -= c
//	a ^= rot(c, 16)
//	c += b
//	b -= a
//	b ^= rot(a, 19)
//	a += c
//	c -= b
//	c ^= rot(b, 4)
//	b += a
//	return c
//}

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
