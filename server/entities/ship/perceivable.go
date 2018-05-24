package ship

import (
	"github.com/20zinnm/spac/server/perceiving"
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/common/net/downstream"
	"math"
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/spac/server/health"
	"github.com/20zinnm/spac/server/shooting"
	"github.com/20zinnm/entity"
)

func Perceivable(
	id entity.ID,
	name string,
	physics world.Component,
	health *health.Component,
	shooting *shooting.Component,
	thrusting *bool,
) perceiving.Perceivable {
	return perceiving.PerceivableFunc(func(b *flatbuffers.Builder, known bool) flatbuffers.UOffsetT {
		var n *flatbuffers.UOffsetT
		if !known {
			n = new(flatbuffers.UOffsetT)
			*n = b.CreateString(name)
		}
		downstream.ShipStart(b)
		downstream.ShipAddPosition(b, downstream.CreateVector(b, float32(physics.Position().X), float32(physics.Position().Y)))
		downstream.ShipAddVelocity(b, downstream.CreateVector(b, float32(physics.Velocity().X), float32(physics.Velocity().Y)))
		downstream.ShipAddAngle(b, float32(physics.Angle()))
		downstream.ShipAddAngularVelocity(b, float32(physics.AngularVelocity()))
		downstream.ShipAddHealth(b, int16(math.Max(float64(*health), 0)))
		if shooting.Armed() {
			downstream.ShipAddArmed(b, 1)
		} else {
			downstream.ShipAddArmed(b, 0)
		}
		if *thrusting {
			downstream.ShipAddThrusting(b, 1)
		} else {
			downstream.ShipAddThrusting(b, 0)
		}
		if n != nil {
			downstream.ShipAddName(b, *n)
		}
		ship := downstream.ShipEnd(b)
		downstream.EntityStart(b)
		downstream.EntityAddId(b, id)
		downstream.EntityAddSnapshotType(b, downstream.SnapshotShip)
		downstream.EntityAddSnapshot(b, ship)
		return downstream.EntityEnd(b)
	})
}
