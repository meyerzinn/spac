package physics

import (
	"github.com/20zinnm/entity"
	"sync"
	"github.com/20zinnm/spac/common/physics/world"
	"github.com/jakecoffman/cp"
)

type System struct {
	world      *world.World
	radius     float64
	manager    *entity.Manager
	entitiesMu sync.RWMutex
	entities   map[entity.ID]Component
}

func New(manager *entity.Manager, world *world.World) *System {
	return &System{
		world:    world,
		entities: make(map[entity.ID]Component),
		manager:  manager,
	}
}

func (s *System) Update(delta float64) {
	s.world.Do(func(space *cp.Space) {
		space.Step(delta)
		s.entitiesMu.RLock()
		for id, component := range s.entities {
			if !component.Position().Near(cp.Vector{}, s.radius) {
				s.manager.Remove(id)
			}
		}
		s.entitiesMu.RUnlock()
	})
}

func (s *System) Add(entity entity.ID, component Component) {
	component.UserData = entity
	s.entitiesMu.Lock()
	s.entities[entity] = component
	s.entitiesMu.Unlock()
}

func (s *System) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	if c, ok := s.entities[entity]; ok {
		s.world.Do(func(space *cp.Space) {
			space.RemoveBody(c.Body)
		})
		delete(s.entities, entity)
	}
	s.entitiesMu.Unlock()
}
