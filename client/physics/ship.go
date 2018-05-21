package physics

import (
	"github.com/20zinnm/entity"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/common/world"
)

var (
	shipVertices = []cp.Vector{{-24, -20}, {24, -20}, {0, 40},}
)

func NewShip(space *cp.Space, id entity.ID) world.Component {
	body := space.AddBody(cp.NewBody(1, cp.MomentForPoly(1, 3, shipVertices, cp.Vector{}, 0)))
	body.UserData = id
	space.AddShape(cp.NewPolyShape(body, 3, shipVertices, cp.NewTransformIdentity(), 0))
	return world.Component{Body: body}
}
