package shooting

import (
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/entity"
)

func NewBullet(space *cp.Space, id entity.ID, owner entity.ID, angle float64, velocity float64) world.Component {
	body := space.AddBody(cp.NewBody(1, cp.MomentForCircle(1, 0, 8, cp.Vector{})))
	body.UserData = id
	bulletShape := space.AddShape(cp.NewCircle(body, 12, cp.Vector{}))
	bulletShape.SetFilter(cp.NewShapeFilter(uint(owner), cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))
	body.SetAngle(angle)
	body.SetVelocityVector(body.Rotation().Rotate(cp.Vector{0, velocity}))
	return world.Component{body}
}
