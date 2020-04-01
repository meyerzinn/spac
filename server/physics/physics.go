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
		body.EachShape(func(shape *cp.Shape) {
			s.space.RemoveShape(shape)
		})
		s.space.RemoveBody(body)
		delete(s.entities, entity)
	}
}

func (s *System) Debug(w io.Writer) {
	fmt.Fprintln(w, "constants system")
	fmt.Fprintf(w, "entities=%d\n", len(s.entities))
	bodies := 0
	s.space.EachBody(func(body *cp.Body) {
		bodies++
	})
	fmt.Fprintf(w, "bodies=%d\n", bodies)
	shapes := 0
	s.space.EachShape(func(shape *cp.Shape) {
		shapes++
	})
	fmt.Fprintf(w, "shapes=%d\n", shapes)
}
