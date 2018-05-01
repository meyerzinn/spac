package rendering

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/20zinnm/entity"
	"github.com/faiface/pixel"
	"github.com/20zinnm/spac/common/physics"
	"sync"
	"image/color"
)

type playingEntity struct {
	*pixel.Sprite
	physics physics.Component
}

type GameRenderer struct {
	window     *pixelgl.Window
	entitiesMu sync.RWMutex
	entities   map[entity.ID]playingEntity
	tracked    entity.ID
	camPos     pixel.Vec
}

func NewGameRenderer(window *pixelgl.Window, tracked entity.ID) *GameRenderer {
	return &GameRenderer{
		window:   window,
		entities: make(map[entity.ID]playingEntity),
		tracked:  tracked,
	}
}

func (s *GameRenderer) Update(delta float64) {
	s.window.Clear(color.White)
	s.entitiesMu.RLock()
	self, ok := s.entities[s.tracked]
	if ok {
		s.camPos = pixel.Vec(self.physics.Position())
	}
	cam := pixel.IM.Scaled(s.camPos, 1).Moved(s.window.Bounds().Center().Sub(s.camPos))
	for _, e := range s.entities {
		posn := e.physics.Position()
		e.Draw(s.window, pixel.IM.Scaled(pixel.ZV, 1).Moved(cam.Unproject(pixel.Vec(posn))).Rotated(pixel.ZV, e.physics.Angle()))
	}
	s.entitiesMu.RUnlock()
}

func (s *GameRenderer) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	delete(s.entities, entity)
	s.entitiesMu.Unlock()
}
