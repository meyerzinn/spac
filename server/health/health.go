package health

import (
	"fmt"
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/server/physics/collision"
	"github.com/jakecoffman/cp"
	"io"
)

type System struct {
	manager  *entity.Manager
	entities map[entity.ID]*Component
}

func New(manager *entity.Manager, space *cp.Space) *System {
	system := &System{
		manager:  manager,
		entities: make(map[entity.ID]*Component),
	}
	/*
		boundary := w.Space.AddBody(cp.NewStaticBody())
		boundaryShape := w.Space.AddShape(cp.NewCircle(boundary, 100, cp.Vector{}))
		boundaryShape.SetSensor(true)
		boundaryShape.SetFilter(cp.NewShapeFilter(0, uint(collision.Boundary), uint(collision.Damageable)))
		boundaryHandler := w.Space.NewCollisionHandler(collision.Boundary, collision.Damageable)
		boundaryHandler.BeginFunc = func(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
			fmt.Println("here c")
			return arb.Ignore()
		}
		boundaryHandler.SeparateFunc = func(arb *cp.Arbiter, space *cp.Space, userData interface{}) {
			system.entitiesMu.RLock()
			defer system.entitiesMu.RUnlock()
			fmt.Println("here a")
			_, bb := arb.Bodies()
			if bid, ok := bb.UserData.(entity.ID); ok {
				if b, ok := system.entities[bid]; ok {
					*b -= 5
					fmt.Println("here b")
				}
			}
		}
	*/
	handler := space.NewCollisionHandler(collision.Damageable, collision.Damageable)
	handler.PreSolveFunc = func(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
		ab, bb := arb.Bodies()
		if aid, ok := ab.UserData.(entity.ID); ok {
			if bid, ok := bb.UserData.(entity.ID); ok {
				if a, ok := system.entities[aid]; ok {
					if b, ok := system.entities[bid]; ok {
						ao, bo := a.Value, b.Value
						a.Value -= bo
						b.Value -= ao
					}
				}
			}
		}
		return true
	}
	return system

}

func (s *System) Add(entity entity.ID, component *Component) {
	s.entities[entity] = component
}

func (s *System) Update(delta float64) {
	for id, entity := range s.entities {
		if entity.Value <= 0 {
			go s.manager.Remove(id)
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	delete(s.entities, entity)
}

func (s *System) Debug(w io.Writer) {
	fmt.Fprintln(w, "health system")
	fmt.Fprintf(w, "count=%d\n", len(s.entities))
	fmt.Fprintln(w, "entities=")
	for id, component := range s.entities {
		fmt.Fprintf(w, "> id=%d max=%f value=%f\n", id, component.Max, component.Value)
	}
}
