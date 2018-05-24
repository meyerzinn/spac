package ship

import (
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/spac/server/physics/collision"
	"github.com/20zinnm/entity"
)

var shipVertices = []cp.Vector{{-24, -20}, {24, -20}, {0, 40},}

func Physics(space *cp.Space, id entity.ID) world.Component {
	body := space.AddBody(cp.NewBody(1, cp.MomentForPoly(1, 3, shipVertices, cp.Vector{}, 0)))
	body.UserData = id
	shipShape := space.AddShape(cp.NewPolyShape(body, 3, shipVertices, cp.NewTransformIdentity(), 0))
	shipShape.SetFilter(cp.NewShapeFilter(uint(id), uint(collision.Health|collision.Perceiving), cp.ALL_CATEGORIES))
	return world.Component{Body: body}
}
