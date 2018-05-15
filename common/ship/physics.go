package ship

import (
	"github.com/20zinnm/spac/common/physics"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/server/perceiving"
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/server/health"
)

var (
	Vertices = []cp.Vector{{-24, -20}, {24, -20}, {0, 40},}
)

func NewPhysics(space *cp.Space, id entity.ID, position cp.Vector) physics.Component {
	body := space.AddBody(cp.NewBody(1, cp.MomentForPoly(1, 3, Vertices, cp.Vector{}, 0)))
	body.UserData = id
	shipShape := space.AddShape(cp.NewPolyShape(body, 3, Vertices, cp.NewTransformIdentity(), 0))
	shipShape.SetFilter(cp.NewShapeFilter(uint(id), uint(perceiving.CollisionType|health.CollisionType), cp.ALL_CATEGORIES))
	body.SetPosition(position)
	return physics.Component{Body: body}
}
