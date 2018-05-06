package rendering

import (
	"github.com/faiface/pixel"
	"github.com/20zinnm/spac/common/physics"
	"github.com/20zinnm/entity"
	"sync"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/imdraw"
	"github.com/20zinnm/spac/common/physics/world"
	"github.com/jakecoffman/cp"
	"math"
	"image/color"
)

type Drawable interface {
	Draw(imd *imdraw.IMDraw)
}

type System struct {
	window  *pixelgl.Window
	world   *world.World
	tracked physics.Component

	//stateMu guards all of the following fields
	stateMu  sync.RWMutex
	entities map[entity.ID]Drawable
	imd      *imdraw.IMDraw
	canvas   *pixelgl.Canvas
	camPos   pixel.Vec
}

func New(window *pixelgl.Window, world *world.World, tracking physics.Component) *System {
	var posn pixel.Vec
	world.Do(func(_ *cp.Space) {
		posn = pixel.Vec(tracking.Position())
	})
	return &System{
		window:   window,
		world:    world,
		tracked:  tracking,
		entities: make(map[entity.ID]Drawable),
		imd:      imdraw.New(nil),
		canvas:   pixelgl.NewCanvas(pixel.R(-1024/2, -768/2, 1024/2, 768/2)),
		camPos:   posn,
	}
}

func (s *System) Add(entity entity.ID, drawable Drawable, physics physics.Component) {
	s.stateMu.Lock()
	s.entities[entity] = drawable
	s.stateMu.Unlock()
}

func (s *System) Update(delta float64) {
	if !s.window.Closed() {
		s.stateMu.RLock()
		defer s.stateMu.RUnlock()

		var posn pixel.Vec
		// synchronize rendering so that we can safely read positions without any random updates
		s.world.Do(func(_ *cp.Space) {
			// Camera
			posn = pixel.Vec(s.tracked.Position())
		})
		s.camPos = pixel.Lerp(s.camPos, posn, 1-math.Pow(1.0/128, delta))
		cam := pixel.IM.Moved(s.camPos.Scaled(-1))
		s.canvas.SetMatrix(cam)

		// Render
		s.canvas.Clear(color.White)
		s.imd.Clear()
		for _, drawable := range s.entities {
			drawable.Draw(s.imd)
		}
		s.imd.Draw(s.canvas)
		s.window.SetMatrix(pixel.IM.Scaled(pixel.ZV,
			math.Min(
				s.window.Bounds().W()/s.canvas.Bounds().W(),
				s.window.Bounds().H()/s.canvas.Bounds().H(),
			),
		).Moved(s.window.Bounds().Center()))
		s.canvas.Draw(s.window, pixel.IM.Moved(s.canvas.Bounds().Center()))
		s.window.Update()
	}
}

func (s *System) Remove(entity entity.ID) {
	s.stateMu.Lock()
	delete(s.entities, entity)
	s.stateMu.Unlock()
}
