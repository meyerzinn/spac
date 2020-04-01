package perceiving

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/client/physics"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/flatbuffers/go"
	"github.com/jakecoffman/cp"
	"image/color"
	"math"
	"sort"
	"time"
)

var (
	shipVertices = []cp.Vector{{0, 51}, {-24, -21}, {0, -9}, {24, -21}}
)

type shipPhysics struct {
	timestamp time.Time
	physics.TranslationalState
	physics.RotationalState
}

func (s shipPhysics) Step(dt float64) shipPhysics {
	return shipPhysics{
		TranslationalState: s.TranslationalState.Step(dt),
		RotationalState:    s.RotationalState.Step(dt),
	}
}

func (s shipPhysics) Lerp(to shipPhysics, delta float64) shipPhysics {
	return shipPhysics{
		TranslationalState: s.TranslationalState.Lerp(to.TranslationalState, delta),
		RotationalState:    s.RotationalState.Lerp(to.RotationalState, delta),
	}
}

type Ship struct {
	shipPhysics
	ID        entity.ID
	health    int
	bufferLen int
	buffer    [InterpolationBuffer]shipPhysics
	Thrusting bool
	Armed     bool
}

func NewShip(id entity.ID) *Ship {
	return &Ship{ID: id}
}

func (s *Ship) Update(timestamp time.Time, table *flatbuffers.Table) {
	shipUpdate := new(downstream.Ship)
	shipUpdate.Init(table.Bytes, table.Pos)
	posn := tocpv(shipUpdate.Position(new(downstream.Vector)))
	vel := tocpv(shipUpdate.Velocity(new(downstream.Vector)))
	angle := float64(shipUpdate.Angle())
	angularVel := float64(shipUpdate.AngularVelocity())
	phys := shipPhysics{
		timestamp: timestamp,
		TranslationalState: physics.TranslationalState{
			Position: posn,
			Velocity: vel,
		},
		RotationalState: physics.RotationalState{
			Angle:           angle,
			AngularVelocity: angularVel,
		},
	}
	copy(s.buffer[1:], s.buffer[:])
	s.buffer[0] = phys
	if s.bufferLen == 0 {
		s.shipPhysics = phys
	}
	if s.bufferLen < InterpolationBuffer {
		s.bufferLen++
	}
	sort.SliceStable(s.buffer[:s.bufferLen], func(i, j int) bool {
		return s.buffer[i].timestamp.After(s.buffer[j].timestamp)
	})
	//// interpolation
	//if posn.DistanceSq(s.lastPos) < 1000 {
	//	fmt.Println("lerping")
	//	s.Physics.SetPosition(s.lastPos.Lerp(posn, delta))
	//	s.Physics.SetVelocityVector(s.lastVel.Lerp(vel, delta))
	//	s.Physics.SetAngle(cp.Lerp(s.lastAngle, angle, delta))
	//	s.Physics.SetAngularVelocity(cp.Lerp(s.lastAngularVel, angularVel, delta))
	//} else {
	//	fmt.Println("jumping")
	//	s.Physics.SetPosition(posn)
	//	s.Physics.SetVelocityVector(vel)
	//	s.Physics.SetAngle(angle)
	//	s.Physics.SetAngularVelocity(angularVel)
	//}
	//s.lastPos = posn
	//s.lastVel = vel
	//s.lastAngle = angle
	//s.lastAngularVel = angularVel
	s.Thrusting = shipUpdate.Thrusting()
	s.Armed = shipUpdate.Armed()
	s.health = int(shipUpdate.Health())
}

func (s *Ship) Position() pixel.Vec {
	return pixel.Vec(s.TranslationalState.Position)
}

func (s *Ship) Health() int {
	return s.health
}

var (
	shipThrusterVertices = []pixel.Vec{{-8, -9}, {8, -9}, {0, -40}}
	shipArmedVertex      = pixel.Vec{Y: 8}
)

func calcLabelY(theta float64) float64 {
	return -12.7096*math.Sin(-2*(theta+3.75912)) + 44
}

func (s *Ship) FixedUpdate() {
	interpolationTime := time.Now().Add(-InterpolationBackTime)
	if s.buffer[0].timestamp.After(interpolationTime) {
		// INTERPOLATE
		for i := 1; i <= s.bufferLen; i++ {
			if s.buffer[i].timestamp.Before(interpolationTime) || i == s.bufferLen-1 {
				newer := s.buffer[i-1]
				older := s.buffer[i]

				length := newer.timestamp.Sub(older.timestamp).Seconds()
				var t float64
				if length > 0.0001 {
					t = interpolationTime.Sub(older.timestamp).Seconds() / length
				}
				s.shipPhysics = s.shipPhysics.Lerp(older.Lerp(newer, t), InterpolationBackTime.Seconds())
				return
			}

		}
	} else {
		// EXTRAPOLATE (rough)
		dt := time.Now().Sub(s.buffer[0].timestamp).Seconds()
		newer := s.shipPhysics.Step(dt)
		s.shipPhysics = s.shipPhysics.Lerp(s.buffer[0].Lerp(newer, InterpolationBackTime.Seconds()), InterpolationConstant)
	}
}

func (s *Ship) Draw(_ *pixelgl.Canvas, imd *imdraw.IMDraw) {

	a := s.Angle
	p := pixel.Vec(s.TranslationalState.Position)
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
	//// draw name
	//if s.Name != "" {
	//	if s.text == nil {
	//		s.text = text.New(pixel.Vec{}, fonts.Atlas)
	//	}
	//	s.text.Clear()
	//	s.text.Write([]byte(s.Name))
	//	s.text.Draw(canvas, pixel.IM.Moved(p.Sub(pixel.Vec{s.text.Bounds().W() / 2, -calcLabelY(s.Physics.Angle())})))
	//	s.text.Clear()
	//	//fmt.Println(s.Physics.Angle(), calcLabelY(s.Physics.Angle()))
	//}
}
