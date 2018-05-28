package bullet

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/server/physics/collision"
	"github.com/jakecoffman/cp"
)

func Physics(space *cp.Space, id entity.ID, owner entity.ID, ownerPhysics *cp.Body, force float64) *cp.Body {
	body := space.AddBody(cp.NewBody(1, cp.MomentForCircle(1, 0, 8, cp.Vector{})))
	body.UserData = id
	bulletShape := space.AddShape(cp.NewCircle(body, 8, cp.Vector{}))
	bulletShape.SetFilter(cp.NewShapeFilter(uint(owner), uint(collision.Damageable|collision.Perceivable), uint(collision.Damageable|collision.Perceiving)))
	body.SetPosition(ownerPhysics.Position().Add(cp.Vector{0, 8}))
	body.SetAngle(ownerPhysics.Angle())
	body.ApplyImpulseAtLocalPoint(cp.Vector{0, force}, cp.Vector{})
	//body.SetVelocityVector(body.Rotation().Rotate(cp.Vector{0, velocity}))
	return body
}
