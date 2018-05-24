package ship

import (
	"github.com/20zinnm/spac/client/rendering"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel"
	"image/color"
	"github.com/20zinnm/spac/common/world"
)

func Renderable(physics world.Component, thrusting *bool, armed *bool) rendering.Renderable {
	return rendering.RenderableFunc(func(imd *imdraw.IMDraw) {
		imd.Color = color.RGBA{
			R: 242,
			G: 75,
			B: 105,
			A: 255,
		}
		a := physics.Angle()
		//p := pixel.Lerp(pixel.Vec(shipPhysics.Position()), lastPosn, 1-math.Pow(1/128, delta))
		//lastPosn = pixel.Vec(shipPhysics.Position())
		p := pixel.Vec(physics.Position())
		imd.Push(
			pixel.Vec{-24, -20}.Rotated(a).Add(p),
			pixel.Vec{24, -20}.Rotated(a).Add(p),
			pixel.Vec{0, 40}.Rotated(a).Add(p),
		)
		imd.Polygon(0)
		if *thrusting {
			imd.Color = color.RGBA{
				R: 235,
				G: 200,
				B: 82,
				A: 255,
			}
			imd.Push(
				pixel.Vec{-8, -20}.Rotated(a).Add(p),
				pixel.Vec{8, -20}.Rotated(a).Add(p),
				pixel.Vec{0, -40}.Rotated(a).Add(p),
			)
			imd.Polygon(0)
		}
		if *armed {
			imd.Color = color.RGBA{
				R: 74,
				G: 136,
				B: 212,
				A: 255,
			}
			imd.Push(p)
			imd.Circle(8, 0)
		}
	})
}
