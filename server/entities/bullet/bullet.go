package bullet

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/server/health"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/server/despawning"
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/server/physics/collision"
)

type Entity struct {
	ID         entity.ID
	Owner      entity.ID
	Health     *health.Component
	Physics    *cp.Body
	Despawning *despawning.Component
}

func New(space *cp.Space, id entity.ID, owner entity.ID, ownerPhysics *cp.Body, force float64, damage float64, ttl uint) *Entity {
	body := space.AddBody(cp.NewBody(1, cp.MomentForCircle(1, 0, 8, cp.Vector{})))
	body.UserData = id
	bulletShape := space.AddShape(cp.NewCircle(body, 8, cp.Vector{}))
	bulletShape.SetCollisionType(collision.Bullet)
	bulletShape.SetFilter(cp.NewShapeFilter(uint(owner), uint(collision.Damageable|collision.Perceivable), cp.ALL_CATEGORIES))
	body.SetPosition(ownerPhysics.Position().Add(cp.Vector{0, 8}.Rotate(ownerPhysics.Rotation())))
	body.SetAngle(ownerPhysics.Angle())
	body.ApplyImpulseAtLocalPoint(cp.Vector{0, force}, cp.Vector{})
	return &Entity{
		ID:    id,
		Owner: owner,
		Health: &health.Component{
			Value: damage,
		},
		Physics: body,
		Despawning: &despawning.Component{
			TTL: ttl,
		},
	}
}

func (e *Entity) Snapshot(b *flatbuffers.Builder, known bool) flatbuffers.UOffsetT {
	downstream.BulletStart(b)
	position := e.Physics.Position()
	downstream.BulletAddPosition(b, downstream.CreateVector(b, float32(position.X), float32(position.Y)))
	velocity := e.Physics.Velocity()
	downstream.BulletAddVelocity(b, downstream.CreateVector(b, float32(velocity.X), float32(velocity.Y)))
	snap := downstream.BulletEnd(b)
	downstream.EntityStart(b)
	downstream.EntityAddId(b, e.ID)
	downstream.EntityAddSnapshot(b, snap)
	downstream.EntityAddSnapshotType(b, downstream.SnapshotBullet)
	return downstream.EntityEnd(b)
}
