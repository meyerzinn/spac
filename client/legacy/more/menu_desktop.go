//+build desktop

package more

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/20zinnm/entity"
)

type MenuRenderer struct {
	window   *pixelgl.Window
	namebuff string
}

func (r *MenuRenderer) Update(delta float64) {
	panic("implement me")
}

func (r *MenuRenderer) Remove(entity entity.ID) {
	panic("implement me")
}

func NewMenuRenderer(window *pixelgl.Window) *MenuRenderer {
	return &MenuRenderer{
		window: window,
	}
}
