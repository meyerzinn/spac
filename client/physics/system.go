package physics

import "github.com/20zinnm/entity"

type Entity interface {
	FixedUpdate()
}

type System struct {
	entities map[entity.ID]Entity
}

func (s *System) Update(float64) {
	for _, e := range s.entities {
		e.FixedUpdate()
	}
}

func (s *System) Add(id entity.ID, e Entity) {
	s.entities[id] = e
}

func (s *System) Remove(e entity.ID) {
	delete(s.entities, e)
}

func New() *System {
	return &System{entities: make(map[entity.ID]Entity)}
}
