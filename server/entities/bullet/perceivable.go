package bullet

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/server/perceiving"
	"github.com/google/flatbuffers/go"
	"github.com/jakecoffman/cp"
)

func Perceivable(id entity.ID, physics *cp.Body) perceiving.Perceivable {
	return perceiving.PerceivableFunc(func(b *flatbuffers.Builder, known bool) flatbuffers.UOffsetT {
		downstream.BulletStart(b)
		position := physics.Position()
		downstream.BulletAddPosition(b, downstream.CreateVector(b, float32(position.X), float32(position.Y)))
		velocity := physics.Velocity()
		downstream.BulletAddVelocity(b, downstream.CreateVector(b, float32(velocity.X), float32(velocity.Y)))
		snap := downstream.BulletEnd(b)
		downstream.EntityStart(b)
		downstream.EntityAddId(b, id)
		downstream.EntityAddSnapshot(b, snap)
		downstream.EntityAddSnapshotType(b, downstream.SnapshotBullet)
		return downstream.EntityEnd(b)
	})
}
