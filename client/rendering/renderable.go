package rendering

import (
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
)

type Renderable interface {
	Draw(canvas *pixelgl.Canvas, imd *imdraw.IMDraw)
}
