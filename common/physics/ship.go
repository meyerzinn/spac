package physics

import (
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/server/perceiving"
	"github.com/20zinnm/entity"
)

var (
	shipVertices = []cp.Vector{{-24, -20}, {24, -20}, {0, 40},}
)

func NewShip(space *cp.Space, id entity.ID) Component {
	c := Component{Body: space.AddBody(cp.NewBody(1, cp.MomentForPoly(1, 3, shipVertices, cp.Vector{}, 0)))}
	shipShape := space.AddShape(cp.NewPolyShape(c.Body, 3, shipVertices, cp.NewTransformIdentity(), 0))
	shipShape.SetFilter(cp.NewShapeFilter(uint(id), uint(perceiving.CollisionType), cp.ALL_CATEGORIES))
	return c
}
