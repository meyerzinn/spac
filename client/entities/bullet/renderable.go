package bullet

import (
	"github.com/20zinnm/spac/client/rendering"
	"github.com/faiface/pixel/imdraw"
	"image/color"
	"github.com/20zinnm/spac/common/world"
	"github.com/faiface/pixel"
)

func Renderable(physics world.Component) rendering.Renderable {
	return rendering.RenderableFunc(func(imd *imdraw.IMDraw) {
		imd.Color = color.RGBA{
			R: 74,
			G: 136,
			B: 212,
			A: 255,
		}
		posn := physics.Position()
		imd.Push(pixel.Vec(posn))
		imd.Circle(8, 0)
	})
}
