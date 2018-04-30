//+build desktop

package rendering

import (
	"github.com/20zinnm/entity"
	"sync"
	"github.com/faiface/pixel/pixelgl"
)

type System struct {
	window     *pixelgl.Window
	entitiesMu sync.RWMutex
	entities   map[entity.ID]Component
}

func New(window *pixelgl.Window) *System {
	system := &System{
		window:   window,
		entities: make(map[entity.ID]Component),
	}
	return system
}

func (s *System) Update(delta float64) {
	if !s.window.Closed() {

	}
}

func (s *System) Remove(entity entity.ID) {
	panic("implement me")
}
