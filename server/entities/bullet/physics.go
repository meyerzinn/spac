package bullet

import (
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/server/physics/collision"
)

func Physics(space *cp.Space, id entity.ID, owner entity.ID, ownerPhysics world.Component, force float64) world.Component {
	body := space.AddBody(cp.NewBody(1, cp.MomentForCircle(1, 0, 8, cp.Vector{})))
	body.UserData = id
	bulletShape := space.AddShape(cp.NewCircle(body, 8, cp.Vector{}))
	bulletShape.SetFilter(cp.NewShapeFilter(uint(owner), uint(collision.Health|collision.Perceiving), cp.ALL_CATEGORIES))
	body.SetPosition(ownerPhysics.Position())
	body.SetAngle(ownerPhysics.Angle())
	body.ApplyImpulseAtLocalPoint(cp.Vector{0, force}, cp.Vector{})
	//body.SetVelocityVector(body.Rotation().Rotate(cp.Vector{0, velocity}))
	return world.Component{Body: body}
}
