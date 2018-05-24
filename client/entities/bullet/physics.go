package bullet

import (
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/common/world"
)

func Physics(space *cp.Space) world.Component {
	body := space.AddBody(cp.NewBody(1, cp.MomentForCircle(1, 0, 8, cp.Vector{})))
	space.AddShape(cp.NewCircle(body, 8, cp.Vector{}))
	return world.Component{Body: body}
}
