package rendering

import (
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type LoadingScene struct {
}

func (s *LoadingScene) Render(window *pixelgl.Window) {
	window.Clear(colornames.Green) // todo make actual connecting screen
}
