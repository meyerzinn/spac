package physics

import (
	"github.com/20zinnm/entity"
	"sync"
	"github.com/20zinnm/spac/common/world"
)

type System struct {
	world      *world.World
	entitiesMu sync.RWMutex
	entities   map[entity.ID]world.Component
}

func New(w *world.World) *System {
	return &System{
		world:    w,
		entities: make(map[entity.ID]world.Component),
	}
}

func (s *System) Update(delta float64) {
	s.world.Lock()
	defer s.world.Unlock()
	s.entitiesMu.RLock()
	defer s.entitiesMu.RUnlock()
	s.world.Space.Step(delta)
}

func (s *System) Add(entity entity.ID, component world.Component) {
	s.entitiesMu.Lock()
	s.entities[entity] = component
	s.entitiesMu.Unlock()
}

func (s *System) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	defer s.entitiesMu.Unlock()
	if c, ok := s.entities[entity]; ok {
		s.world.Lock()
		s.world.Space.RemoveBody(c.Body)
		s.world.Unlock()
		delete(s.entities, entity)
	}
}
