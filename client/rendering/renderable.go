package rendering

import "github.com/faiface/pixel/imdraw"

type Renderable interface {
	Draw(imd *imdraw.IMDraw)
}

type RenderableFunc func(imd *imdraw.IMDraw)

func (fn RenderableFunc) Draw(imd *imdraw.IMDraw) {
	fn(imd)
}
