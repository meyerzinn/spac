package ship

import (
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/google/flatbuffers/go"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/common/world"
)

func Updater(physics world.Component, thrusting *bool, armed *bool) func([]byte, flatbuffers.UOffsetT) {
	return func(bytes []byte, pos flatbuffers.UOffsetT) {
		shipUpdate := new(downstream.Ship)
		shipUpdate.Init(bytes, pos)
		posn := shipUpdate.Position(new(downstream.Vector))
		vel := shipUpdate.Velocity(new(downstream.Vector))
		physics.SetPosition(cp.Vector{X: float64(posn.X()), Y: float64(posn.Y())})
		physics.SetVelocity(float64(vel.X()), float64(vel.Y()))
		physics.SetAngle(float64(shipUpdate.Angle()))
		physics.SetAngularVelocity(float64(shipUpdate.AngularVelocity()))
		*thrusting = shipUpdate.Thrusting() > 0
		*armed = shipUpdate.Armed() > 0
	}
}
