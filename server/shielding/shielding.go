package shielding

import (
	"github.com/20zinnm/spac/server/health"
	"github.com/20zinnm/entity"
	"github.com/jakecoffman/cp"
)

type Component struct {
	// Health is the amount of health the shield has.
	Health *health.Component
	// Max is the maximum amount of health that the shield can have.
	Max float64
	// Min is the minimum amount of health needed to activate the shield (though once active, health can be drained to zero).
	Min float64
	// Recharge is the amount of health the shield regains per tick of inactivity.
	Recharge float64
	// Delay is the number of ticks between the the player signalling for the shield to activate and it activating.
	// The shield can be cancelled during the activation sequence but it will still consume health.
	Delay int
	// Decay is the minimum amount of health the shield loses per tick.
	Decay float64
}

type Controller chan bool

type shieldingEntity struct {
	controller Controller
	component  *Component
	// activation is the number of ticks since the user signalled for their shield.
	activation int
	physics    *cp.Body
	shield     *cp.Shape
	shielding  bool
	active     bool
}

type System struct {
	space    *cp.Space
	entities map[entity.ID]shieldingEntity
}

func (s *System) Update(delta float64) {
	for _, e := range s.entities {
		select {
		case n := <-e.controller:
			if e.shielding != n {
				if e.active {
				}
			}
		default:
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	panic("implement me")
}
