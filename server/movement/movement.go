package movement

import (
	"github.com/20zinnm/spac/common/physics"
	"github.com/20zinnm/entity"
	"sync"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/common/physics/world"
	"fmt"
)

type movementEntity struct {
	controller Controller
	physics    physics.Component
	//control    physics.Component
	move       float64
	turn       float64
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
	for _, e := range s.entities {
		controls := e.controller.Controls()
		if controls.Left != controls.Right {
			if controls.Left {
				s.world.Do(func(_ *cp.Space) {
					e.physics.SetAngularVelocity(e.physics.Angle() + e.turn)
				})
			} else {
				s.world.Do(func(_ *cp.Space) {
					e.physics.SetAngularVelocity(e.physics.Angle() - e.turn)
				})
			}
		}
		if controls.Thrusting {
			s.world.Do(func(_ *cp.Space) {
				fmt.Println("applying thrust", e.physics.Position(), e.physics.Velocity(), e.physics.Force())
				e.physics.ApplyForceAtLocalPoint(cp.Vector{Y: e.move}, cp.Vector{})
			})
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	if _, ok := s.entities[entity]; ok {
		//s.world.Do(func(space *cp.Space) {
			//space.RemoveBody(e.control.Body)
		//})
		delete(s.entities, entity)
	}
	s.entitiesMu.Unlock()
}

func (s *System) Add(id entity.ID, controller Controller, body physics.Component, move float64, turn float64) {
	//var controlBody *cp.Body
	//var pivot *cp.Constraint
	//var gear *cp.Constraint
	//s.world.Do(func(space *cp.Space) {
		//controlBody = space.AddBody(cp.NewKinematicBody())
		//pivot = space.AddConstraint(cp.NewPivotJoint2(controlBody, body.Body, cp.Vector{}, cp.Vector{}))
		//gear = space.AddConstraint(cp.NewGearJoint(controlBody, body.Body, 0.0, 1.0))
		//
		//pivot.SetMaxBias(0)
		//pivot.SetMaxForce(10000)

		//gear.SetErrorBias(0)    // attempt to fully correct the joint each step
		//gear.SetMaxBias(1.2)    // but limit it's angular correction rate
		//gear.SetMaxForce(50000) // emulate angular friction
	//})
	s.entitiesMu.Lock()
	s.entities[id] = movementEntity{
		controller: controller,
		physics:    body,
		//control:    physics.Component{Body: controlBody},
		move:       move,
		turn:       turn,
	}
	s.entitiesMu.Unlock()
}
