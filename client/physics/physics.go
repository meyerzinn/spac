package physics

import (
	"github.com/20zinnm/spac/common/constants"
	"github.com/jakecoffman/cp"
)

type TranslationalState struct {
	Position cp.Vector
	Velocity cp.Vector
}

func (t TranslationalState) Step(dt float64) TranslationalState {
	return TranslationalState{
		Position: t.Position.Add(t.Velocity.Mult(dt)),
		Velocity: t.Velocity.Mult(constants.Damping),
	}
}

func (t TranslationalState) Lerp(to TranslationalState, delta float64) TranslationalState {
	return TranslationalState{
		Position: t.Position.Lerp(to.Position, delta),
		Velocity: t.Velocity.Lerp(to.Velocity, delta),
	}
}

type RotationalState struct {
	Angle           float64
	AngularVelocity float64
}

func (r RotationalState) Step(dt float64) RotationalState {
	return RotationalState{
		Angle:           r.Angle + r.AngularVelocity * dt,
		AngularVelocity: r.AngularVelocity * (constants.Damping),
	}
}

func (r RotationalState) Lerp(to RotationalState, delta float64) RotationalState {
	return RotationalState{
		Angle:           cp.Lerp(r.Angle, to.Angle, delta),
		AngularVelocity: cp.Lerp(r.AngularVelocity, to.AngularVelocity, delta),
	}
}
