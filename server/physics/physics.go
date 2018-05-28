package physics

import (
	"fmt"
	"github.com/20zinnm/entity"
	"github.com/jakecoffman/cp"
	"io"
)

type System struct {
	space    *cp.Space
	manager  *entity.Manager
	entities map[entity.ID]*cp.Body
}

func New(manager *entity.Manager, space *cp.Space) *System {
	return &System{
		space:    space,
		entities: make(map[entity.ID]*cp.Body),
		manager:  manager,
	}
}

func (s *System) Update(delta float64) {
	s.space.Step(delta)
}

func (s *System) Add(entity entity.ID, body *cp.Body) {
	s.entities[entity] = body
	body.UserData = entity
}

func (s *System) Remove(entity entity.ID) {
	if body, ok := s.entities[entity]; ok {
		s.space.RemoveBody(body)
		delete(s.entities, entity)
	}
}

func (s *System) Debug(w io.Writer) {
	fmt.Fprintln(w, "physics system")
	fmt.Fprintf(w, "entities=%d\n", len(s.entities))
}
