package rendering

import (
	"github.com/faiface/pixel/imdraw"
	"image/color"
	"github.com/faiface/pixel"
	"github.com/20zinnm/spac/common/physics"
	"sync/atomic"
	"github.com/20zinnm/spac/common/physics/world"
	"github.com/jakecoffman/cp"
)

type ship struct {
	physics   physics.Component
	thrusting *int32
	armed     *int32
	world     *world.World
}

func Ship(physics physics.Component, thrusting *int32, armed *int32, world *world.World) Drawable {
	return &ship{
		physics:   physics,
		thrusting: thrusting,
		armed:     armed,
		world:     world,
	}
}

func (s *ship) Draw(imd *imdraw.IMDraw) {
	imd.Color = color.RGBA{
		R: 242,
		G: 75,
		B: 105,
		A: 1,
	}
	var posn pixel.Vec
	s.world.Do(func(_ *cp.Space) {
		posn = pixel.Vec(s.physics.Position())
	})
	imd.Push(
		pixel.Vec{-24, -20}.Add(posn),
		pixel.Vec{24, -20}.Add(posn),
		pixel.Vec{0, 40}.Add(posn),
	)
	imd.Polygon(0)
	if atomic.LoadInt32(s.thrusting) > 0 {
		imd.Color = color.RGBA{
			R: 235,
			G: 200,
			B: 82,
			A: 1,
		}
		imd.Push(
			pixel.Vec{-8, -20}.Add(posn),
			pixel.Vec{8, -20}.Add(posn),
			pixel.Vec{0, -40}.Add(posn),
		)
		imd.Polygon(0)
	}
	if atomic.LoadInt32(s.armed) > 0 {
		imd.Color = color.RGBA{
			R: 74,
			G: 136,
			B: 212,
			A: 1,
		}
		imd.Push(pixel.ZV.Add(posn))
		imd.Circle(8, 0)
	}
	//pixel.IM.Scaled(pixel.ZV, 1).Moved(cam.Unproject(pixel.Vec{X: posn.X, Y: posn.Y
}
