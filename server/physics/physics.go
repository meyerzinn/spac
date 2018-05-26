package physics

import (
	"github.com/20zinnm/entity"
	"sync"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/common/world"
)

type System struct {
	world      *world.World
	radius     float64
	manager    *entity.Manager
	entitiesMu sync.RWMutex
	entities   map[entity.ID]world.Component
}

func New(manager *entity.Manager, w *world.World, radius float64) *System {
	return &System{
		world:    w,
		radius:   radius,
		entities: make(map[entity.ID]world.Component),
		manager:  manager,
	}
}

func (s *System) Update(delta float64) {
	s.world.Lock()
	defer s.world.Unlock()

	s.world.Space.Step(delta)
	s.entitiesMu.RLock()
	for id, component := range s.entities {
		if !component.Position().Near(cp.Vector{}, s.radius) {
			go s.manager.Remove(id)
		}
	}
	s.entitiesMu.RUnlock()
}

func (s *System) Add(entity entity.ID, component world.Component) {
	s.entitiesMu.Lock()
	s.entities[entity] = component
	s.entitiesMu.Unlock()
}

func (s *System) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	if c, ok := s.entities[entity]; ok {
		s.world.Lock()
		s.world.Space.RemoveBody(c.Body)
		s.world.Unlock()
		delete(s.entities, entity)
	}
	s.entitiesMu.Unlock()
}
