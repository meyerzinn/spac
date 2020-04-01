package movement

import (
	"fmt"
	"github.com/20zinnm/entity"
	"github.com/jakecoffman/cp"
	"io"
)

type Controller chan Controls

type movementEntity struct {
	in      Controller
	last    Controls
	physics *cp.Body
	linear  float64
	angular float64
}

type System struct {
	entities map[entity.ID]*movementEntity
}

func New() *System {
	return &System{
		entities: make(map[entity.ID]*movementEntity),
	}
}

func (s *System) Update(delta float64) {
	for _, e := range s.entities {
		select {
		case n := <-e.in:
			e.last = n
		default:
		}
		if e.last.Left != e.last.Right {
			if e.last.Left {
				e.physics.ApplyForceAtLocalPoint(cp.Vector{X: -e.angular}, cp.Vector{Y: 1})
				e.physics.ApplyForceAtLocalPoint(cp.Vector{X: e.angular}, cp.Vector{})
				//e.physics.SetAngularVelocity(e.angular)
			} else {
				e.physics.ApplyForceAtLocalPoint(cp.Vector{X: e.angular}, cp.Vector{Y: 1})
				e.physics.ApplyForceAtLocalPoint(cp.Vector{X: -e.angular}, cp.Vector{})
				//e.physics.SetAngularVelocity(-e.angular)
			}
		}
		if e.last.Thrusting {
			e.physics.SetForce(e.physics.Rotation().Rotate(cp.Vector{Y: e.linear}))
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	if _, ok := s.entities[entity]; ok {
		delete(s.entities, entity)
	}
}

func (s *System) Add(id entity.ID, controller Controller, physics *cp.Body, linearForce float64, angularVelocity float64) {
	s.entities[id] = &movementEntity{
		in:      controller,
		physics: physics,
		linear:  linearForce,
		angular: angularVelocity,
	}
}

func (s *System) Debug(w io.Writer) {
	fmt.Fprintln(w, "movement system")
	fmt.Fprintf(w, "count=%d\n", len(s.entities))
	fmt.Fprintln(w, "entities=")
	for id, e := range s.entities {
		fmt.Fprintf(w, "> id=%d component=%v\n", id, *e)
	}
}
