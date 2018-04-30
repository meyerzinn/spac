//+build desktop

package rendering

import (
	"github.com/20zinnm/entity"
	"sync"
	"github.com/faiface/pixel/pixelgl"
	"fmt"
	"time"
	"github.com/faiface/pixel"
	"github.com/20zinnm/spac/common/physics"
	"image/color"
)

type System struct {
	windowMu   sync.Mutex
	window     *pixelgl.Window
	entitiesMu sync.RWMutex
	entities   map[entity.ID]Component
	tracked    physics.Component
	frames     int
	second     <-chan time.Time
	zoom       float64
}

func New(window *pixelgl.Window, tracked physics.Component) *System {
	system := &System{
		window:   window,
		entities: make(map[entity.ID]Component),
		tracked:  tracked,
		zoom:     1.0,
	}
	return system
}

func (s *System) Update(delta float64) {
	s.windowMu.Lock()
	defer s.windowMu.Unlock()
	if !s.window.Closed() {
		s.window.Update()

		// FPS
		s.frames++
		select {
		case <-s.second:
			s.window.SetTitle(fmt.Sprintf("%s | FPS: %d", "spac", s.frames))
			s.frames = 0
		}

		// Camera
		posn := s.tracked.Position()
		camPos := pixel.Vec{X: posn.X, Y: posn.Y}
		cam := pixel.IM.Scaled(camPos, s.zoom).Moved(s.window.Bounds().Center().Sub(camPos))
		s.window.SetMatrix(cam)

		// Render
		s.window.Clear(color.White)
		s.entitiesMu.RLock()
		for _, component := range s.entities {
			posn := component.Physics.Position()
			component.Draw(s.window, pixel.IM.Scaled(pixel.ZV, s.zoom).Moved(cam.Unproject(pixel.Vec{X: posn.X, Y: posn.Y})))
		}
		s.entitiesMu.RUnlock()
	}
}

func (s *System) Remove(entity entity.ID) {
	panic("implement me")
}
