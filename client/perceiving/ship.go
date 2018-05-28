package perceiving

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/client/fonts"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/google/flatbuffers/go"
	"github.com/jakecoffman/cp"
	"image/color"
	"math"
	"sync"
)

var (
	shipVertices = []cp.Vector{{0, 51}, {-24, -21}, {0, -9}, {24, -21}}
	textPool     = &sync.Pool{
		New: func() interface{} {
			return text.New(pixel.ZV, fonts.Atlas)
		},
	}
)

type Ship struct {
	ID        entity.ID
	health    int
	Physics   *cp.Body
	Thrusting bool
	Armed     bool
	Name      string
}

func NewShip(space *cp.Space, id entity.ID) *Ship {
	body := space.AddBody(cp.NewBody(1, cp.MomentForPoly(1, 3, shipVertices, cp.Vector{}, 0)))
	space.AddShape(cp.NewPolyShape(body, 3, shipVertices, cp.NewTransformIdentity(), 0))

	return &Ship{
		ID:        id,
		Physics:   body,
		Thrusting: false,
		Armed:     false,
	}
}
func (s *Ship) Update(table *flatbuffers.Table) {
	shipUpdate := new(downstream.Ship)
	shipUpdate.Init(table.Bytes, table.Pos)
	posn := shipUpdate.Position(new(downstream.Vector))
	vel := shipUpdate.Velocity(new(downstream.Vector))
	if shipUpdate.Name() != nil {
		s.Name = string(shipUpdate.Name())
	}
	//fmt.Println(posn.X(), posn.Y())
	s.Physics.SetPosition(cp.Vector{X: float64(posn.X()), Y: float64(posn.Y())})
	s.Physics.SetVelocity(float64(vel.X()), float64(vel.Y()))
	s.Physics.SetAngle(float64(shipUpdate.Angle()))
	s.Physics.SetAngularVelocity(float64(shipUpdate.AngularVelocity()))
	s.Thrusting = shipUpdate.Thrusting() > 0
	s.Armed = shipUpdate.Armed() > 0
	s.health = int(shipUpdate.Health())

}

func (s *Ship) Position() pixel.Vec {
	return pixel.Vec(s.Physics.Position())
}

func (s *Ship) Health() int {
	return s.health
}

var (
	shipThrusterVertices = []pixel.Vec{{-8, -9}, {8, -9}, {0, -40}}
	shipArmedVertex      = pixel.Vec{0, 8}
)

func calcLabelY(theta float64) float64 {
	return -12.7096*math.Sin(-2*(theta+3.75912)) + 44
}

func (s *Ship) Draw(canvas *pixelgl.Canvas, imd *imdraw.IMDraw) {
	a := s.Physics.Angle()
	p := pixel.Vec(s.Physics.Position())
	// draw thruster
	if s.Thrusting {
		imd.Color = color.RGBA{
			R: 248,
			G: 196,
			B: 69,
			A: 255,
		}
		for _, v := range shipThrusterVertices {
			imd.Push(v.Rotated(a).Add(p))
		}
		imd.Polygon(0)
	}
	// draw body
	imd.Color = color.RGBA{
		R: 242,
		G: 75,
		B: 105,
		A: 255,
	}
	for _, v := range shipVertices {
		imd.Push(pixel.Vec(v).Rotated(a).Add(p))
	}
	imd.Polygon(0)
	// draw bullet
	if s.Armed {
		imd.Color = color.RGBA{
			R: 74,
			G: 136,
			B: 212,
			A: 255,
		}
		imd.Push(shipArmedVertex.Rotated(a).Add(p))
		imd.Circle(8, 0)
	}
	// draw name
	if s.Name != "" {
		txt := textPool.Get().(*text.Text)
		defer textPool.Put(txt)
		txt.Clear()
		txt.Write([]byte(s.Name))
		txt.Draw(canvas, pixel.IM.Moved(p.Sub(pixel.Vec{txt.Bounds().W() / 2, -calcLabelY(s.Physics.Angle())})))
		txt.Clear()
		//fmt.Println(s.Physics.Angle(), calcLabelY(s.Physics.Angle()))
	}
}
