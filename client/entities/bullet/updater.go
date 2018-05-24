package bullet

import (
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/common/world"
	"github.com/google/flatbuffers/go"
	"github.com/jakecoffman/cp"
)

func Updater(physics world.Component) func([]byte, flatbuffers.UOffsetT) {
	return func(bytes []byte, pos flatbuffers.UOffsetT) {
		bullet := new(downstream.Bullet)
		bullet.Init(bytes, pos)
		posn := bullet.Position(new(downstream.Vector))
		vel := bullet.Velocity(new(downstream.Vector))
		physics.SetPosition(cp.Vector{X: float64(posn.X()), Y: float64(posn.Y())})
		physics.SetVelocityVector(cp.Vector{X: float64(vel.X()), Y: float64(vel.Y())})
	}
}
