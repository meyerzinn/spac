package rendering

import (
	"github.com/faiface/pixel/pixelgl"
)

type Scene interface {
	Render(window *pixelgl.Window)
}
