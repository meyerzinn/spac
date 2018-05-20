package movement

import (
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/entity"
	"sync"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/server/physics"
)

type movementEntity struct {
	controller Controller
	physics    physics.Component
	move float64
	turn float64
}

type System struct {
	entitiesMu sync.RWMutex
	entities   map[entity.ID]movementEntity
	world      *world.World
}

func New(world *world.World) *System {
	return &System{
		world:    world,
		entities: make(map[entity.ID]movementEntity),
	}
}

func (s *System) Update(delta float64) {
	s.entitiesMu.RLock()
	defer s.entitiesMu.RUnlock()
	s.world.Lock()
	defer s.world.Unlock()
	for _, e := range s.entities {
		controls := e.controller.Controls()
		if controls.Left != controls.Right {
			if controls.Left {
				e.physics.SetAngularVelocity(e.turn)
			} else {
				e.physics.SetAngularVelocity(-e.turn)
			}
		} else {
			e.physics.SetAngularVelocity(0)
		}
		if controls.Thrusting {
			e.physics.SetForce(e.physics.Rotation().Rotate(cp.Vector{Y: e.move}))
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

func (s *System) Add(id entity.ID, controller Controller, component physics.Component, move float64, turn float64) {
	s.entitiesMu.Lock()
	s.entities[id] = movementEntity{
		controller: controller,
		physics:    component,
		move: move,
		turn: turn,
	}
	s.entitiesMu.Unlock()
}
