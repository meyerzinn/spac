package movement

import (
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/entity"
	"sync"
	"github.com/jakecoffman/cp"
	"io"
	"fmt"
)

type Controller chan Controls

type movementEntity struct {
	in      Controller
	last    Controls
	physics world.Component
	linear  float64
	angular float64
}

type System struct {
	entitiesMu sync.RWMutex
	entities   map[entity.ID]*movementEntity
	world      *world.World
}

func New(world *world.World) *System {
	return &System{
		world:    world,
		entities: make(map[entity.ID]*movementEntity),
	}
}

func (s *System) Update(delta float64) {
	s.entitiesMu.RLock()
	defer s.entitiesMu.RUnlock()
	s.world.Lock()
	defer s.world.Unlock()
	for _, e := range s.entities {
		select {
		case n := <-e.in:
			e.last = n
		default:
		}
		if e.last.Left != e.last.Right {
			if e.last.Left {
				e.physics.SetAngularVelocity(e.angular)
			} else {
				e.physics.SetAngularVelocity(-e.angular)
			}
		}
		if e.last.Thrusting {
			e.physics.SetForce(e.physics.Rotation().Rotate(cp.Vector{Y: e.linear}))
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	if _, ok := s.entities[entity]; ok {
		delete(s.entities, entity)
	}
	s.entitiesMu.Unlock()
}

func (s *System) Add(id entity.ID, controller Controller, physics world.Component, linearForce float64, angularVelocity float64) {
	s.entitiesMu.Lock()
	s.entities[id] = &movementEntity{
		in:      controller,
		physics: physics,
		linear:  linearForce,
		angular: angularVelocity,
	}
	s.entitiesMu.Unlock()
}

func (s *System) Debug(w io.Writer) {
	s.entitiesMu.Lock()
	defer s.entitiesMu.Unlock()
	fmt.Fprintln(w, "movement system")
	fmt.Fprintf(w, "count=%d\n", len(s.entities))
	fmt.Fprintln(w, "entities=")
	for id, e := range s.entities {
		fmt.Fprintf(w, "> id=%d component=%v\n", id, *e)
	}
}
