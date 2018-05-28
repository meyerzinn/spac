package physics

import (
	"github.com/20zinnm/entity"
	"github.com/jakecoffman/cp"
)

type System struct {
	space    *cp.Space
	entities map[entity.ID]*cp.Body
}

func New(space *cp.Space) *System {
	return &System{
		space:    space,
		entities: make(map[entity.ID]*cp.Body),
	}
}

func (s *System) Update(delta float64) {
	s.space.Step(delta)
}

func (s *System) Add(entity entity.ID, component *cp.Body) {
	s.entities[entity] = component
}

func (s *System) Remove(entity entity.ID) {
	if body, ok := s.entities[entity]; ok {
		s.space.RemoveBody(body)
		delete(s.entities, entity)
	}
}
