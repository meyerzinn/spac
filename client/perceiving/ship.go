package perceiving

import (
	"github.com/20zinnm/spac/common/world"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/entity"
	"github.com/faiface/pixel/imdraw"
	"image/color"
	"github.com/faiface/pixel"
	"sync"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/google/flatbuffers/go"
)

var (
	shipVertices = []cp.Vector{{-24, -20}, {24, -20}, {0, 40},}
)

type Ship struct {
	ID        entity.ID
	Physics   world.Component
	Thrusting bool
	Armed     bool
	sync.RWMutex
}

func NewShip(space *cp.Space, id entity.ID) *Ship {
	body := space.AddBody(cp.NewBody(1, cp.MomentForPoly(1, 3, shipVertices, cp.Vector{}, 0)))
	space.AddShape(cp.NewPolyShape(body, 3, shipVertices, cp.NewTransformIdentity(), 0))
	physics := world.Component{Body: body}

	return &Ship{
		ID:        id,
		Physics:   physics,
		Thrusting: false,
		Armed:     false,
	}
}
func (s *Ship) Update(bytes []byte, pos flatbuffers.UOffsetT) {
	shipUpdate := new(downstream.Ship)
	shipUpdate.Init(bytes, pos)
	posn := shipUpdate.Position(new(downstream.Vector))
	vel := shipUpdate.Velocity(new(downstream.Vector))
	s.Lock()
	defer s.Unlock()
	s.Physics.SetPosition(cp.Vector{X: float64(posn.X()), Y: float64(posn.Y())})
	s.Physics.SetVelocity(float64(vel.X()), float64(vel.Y()))
	s.Physics.SetAngle(float64(shipUpdate.Angle()))
	s.Physics.SetAngularVelocity(float64(shipUpdate.AngularVelocity()))
	s.Thrusting = shipUpdate.Thrusting() > 0
	s.Armed = shipUpdate.Armed() > 0

}

func (s *Ship) Position() pixel.Vec {
	return pixel.Vec(s.Physics.Position())
}

func (s *Ship) Draw(imd *imdraw.IMDraw) {
	s.RLock()
	defer s.RUnlock()

	imd.Color = color.RGBA{
		R: 242,
		G: 75,
		B: 105,
		A: 255,
	}
	a := s.Physics.Angle()
	//p := pixel.Lerp(pixel.Vec(shipPhysics.Position()), lastPosn, 1-math.Pow(1/128, delta))
	//lastPosn = pixel.Vec(shipPhysics.Position())
	p := pixel.Vec(s.Physics.Position())
	imd.Push(
		pixel.Vec{-24, -20}.Rotated(a).Add(p),
		pixel.Vec{24, -20}.Rotated(a).Add(p),
		pixel.Vec{0, 40}.Rotated(a).Add(p),
	)
	imd.Polygon(0)
	if s.Thrusting {
		imd.Color = color.RGBA{
			R: 235,
			G: 200,
			B: 82,
			A: 255,
		}
		imd.Push(
			pixel.Vec{-8, -20}.Rotated(a).Add(p),
			pixel.Vec{8, -20}.Rotated(a).Add(p),
			pixel.Vec{0, -40}.Rotated(a).Add(p),
		)
		imd.Polygon(0)
	}
	if s.Armed {
		imd.Color = color.RGBA{
			R: 74,
			G: 136,
			B: 212,
			A: 255,
		}
		imd.Push(p)
		imd.Circle(8, 0)
	}
}
