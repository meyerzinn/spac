//+build desktop

package rendering

import (
	"github.com/20zinnm/entity"
	"github.com/faiface/pixel/pixelgl"
	"fmt"
	"time"
	"os"
)

//type renderingEntity struct {
//	Component
//	Physics physics.Component
//}

type System struct {
	window *pixelgl.Window
	//entitiesMu sync.RWMutex
	//entities   map[entity.ID]renderingEntity
	//tracked    physics.Component
	frames int
	second <-chan time.Time
	//zoom       float64
}

func New(window *pixelgl.Window) *System {
	system := &System{
		window: window,
		//entities: make(map[entity.ID]renderingEntity),
		//zoom:     1.0,
		//second:   time.Tick(time.Second),
		//tracked:  tracking,
	}
	return system
}

func (s *System) Update(delta float64) {
	if s.window.Closed() {
		os.Exit(0)
	}
	s.window.Update()
	// FPS
	s.frames++
	select {
	case <-s.second:
		s.window.SetTitle(fmt.Sprintf("%s | FPS: %d", "spac", s.frames))
		s.frames = 0
	}

	//// Camera
	//posn := s.tracked.Position()
	//camPos := pixel.Vec{X: posn.X, Y: posn.Y}
	//cam := pixel.IM.Scaled(camPos, s.zoom).Moved(s.window.Bounds().Center().Sub(camPos))
	//s.window.SetMatrix(cam)
	//
	//// Render
	//s.window.Clear(color.White)
	//s.entitiesMu.RLock()
	//for _, component := range s.entities {
	//	posn := component.Physics.Position()
	//	component.Draw(s.window, pixel.IM.Scaled(pixel.ZV, s.zoom).Moved(cam.Unproject(pixel.Vec{X: posn.X, Y: posn.Y})).Rotated(pixel.ZV, component.Physics.Angle()))
	//}
	//s.entitiesMu.RUnlock()
}

func (s *System) Remove(entity entity.ID) {
	panic("implement me")
}
