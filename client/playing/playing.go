package playing

import (
	"github.com/faiface/pixel/pixelgl"
	"image/color"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/common/physics"
	"github.com/faiface/pixel"
	"sync/atomic"
	"github.com/google/flatbuffers/go"
	"github.com/faiface/pixel/imdraw"
	"github.com/20zinnm/entity"
	"golang.org/x/image/colornames"
	"time"
	"github.com/20zinnm/spac/common/ship"
	"math"
	"log"
)

type Handler interface {
	Handle(Controls)
}

type HandlerFunc func(Controls)

func (fn HandlerFunc) Handle(c Controls) {
	fn(c)
}

type Controls struct {
	Left, Right, Thrust, Shoot bool
}

type Entity interface {
	Update(*downstream.Entity)
	Draw(imd *imdraw.IMDraw)
}

type Camera interface {
	Position() pixel.Vec
}

type Scene struct {
	win            *pixelgl.Window
	entities       map[entity.ID]Entity
	world          *physics.World
	canvas         *pixelgl.Canvas
	imd            *imdraw.IMDraw
	camPos         pixel.Vec
	target         *shipEntity // can be nil!
	targetId       entity.ID
	handler        Handler
	lastPerception int64
}

func New(win *pixelgl.Window, worldRadius float64, targetId entity.ID, handler Handler) *Scene {
	world := &physics.World{
		Space: physics.NewSpace(),
	}

	return &Scene{
		win:      win,
		world:    world,
		entities: make(map[entity.ID]Entity),
		//canvas:   pixelgl.NewCanvas(pixel.R(-1024/2, -768/2, 1024/2, 768/2)),
		canvas:   pixelgl.NewCanvas(pixel.R(-1024/2, -768/2, 1024/2, 768/2)),
		imd:      imdraw.New(nil),
		camPos:   pixel.ZV,
		targetId: targetId,
		handler:  handler,
	}
}

func (s *Scene) Perceive(perception *downstream.Perception) {
	atomic.StoreInt64(&s.lastPerception, int64(time.Now().Nanosecond()))
	s.world.Lock()
	defer s.world.Unlock()
	known := make(map[entity.ID]struct{}, perception.EntitiesLength())
	for i := 0; i < perception.EntitiesLength(); i++ {
		e := new(downstream.Entity)
		if !perception.Entities(e, i) {
			continue
		}
		known[e.Id()] = struct{}{}
		updater, ok := s.entities[e.Id()]
		if ok {
			updater.Update(e)
			continue
		}
		switch e.SnapshotType() {
		case downstream.SnapshotShip:
			newShip := new(shipEntity)
			newShip.id = e.Id()
			newShip.physics = ship.NewPhysics(s.world.Space, e.Id(), cp.Vector{})
			newShip.Update(e)
			s.entities[e.Id()] = newShip

			if s.target == nil && e.Id() == s.targetId {
				log.Print("tracking ", e.Id())
				s.target = newShip
			}
		}
	}
	for id := range s.entities {
		if _, ok := known[id]; !ok {
			delete(s.entities, id)
		}
	}
}

func (s *Scene) Update(dt float64) {
	if s.target == nil {
		return
	}
	s.world.Lock()
	defer s.world.Unlock()
	controls := Controls{
		Left:   s.win.Pressed(pixelgl.KeyA),
		Right:  s.win.Pressed(pixelgl.KeyD),
		Thrust: s.win.Pressed(pixelgl.KeyW),
		Shoot:  s.win.Pressed(pixelgl.MouseButton1) || s.win.Pressed(pixelgl.KeySpace),
	}
	// todo implement client-side prediction (this code makes it move too fast--need to investigate)
	//if controls.Left != controls.Right {
	//	if controls.Left {
	//		s.target.physics.SetAngularVelocity(1)
	//	} else {
	//		s.target.physics.SetAngularVelocity(-1)
	//	}
	//} else {
	//	s.target.physics.SetAngularVelocity(0)
	//}
	//if controls.Thrust {
	//	s.target.physics.SetForce(s.target.physics.Rotation().Rotate(cp.Vector{Y: 100}))
	//}
	physicsDelta := time.Now().Sub(time.Unix(0, atomic.LoadInt64(&s.lastPerception))).Seconds()
	s.world.Space.Step(physicsDelta)
	s.handler.Handle(controls)
	if s.win.Bounds() != s.canvas.Bounds() {
		//todo resize canvas if the window size changes
	}
	s.camPos = pixel.Lerp(s.camPos, s.target.Position(), 1-math.Pow(1.0/128, dt))
	cam := pixel.IM.Moved(s.camPos.Scaled(-1))
	s.canvas.Clear(colornames.Black)
	s.imd.Clear()
	s.canvas.SetMatrix(cam)

	// draw stars
	s.imd.Color = colornames.White
	drawStars(s.imd, s.camPos, s.canvas.Bounds(), 4)

	// draw entities
	for _, renderable := range s.entities {
		renderable.Draw(s.imd)
	}
	s.imd.Draw(s.canvas)

	s.win.Clear(colornames.White)
	s.win.SetMatrix(pixel.IM.Scaled(pixel.ZV,
		math.Max(
			s.win.Bounds().W()/s.canvas.Bounds().W(),
			s.win.Bounds().H()/s.canvas.Bounds().H(),
		),
	).Moved(s.win.Bounds().Center()))
	s.canvas.Draw(s.win, pixel.IM.Moved(s.canvas.Bounds().Center()))
}

type shipEntity struct {
	id        entity.ID
	physics   physics.Component
	thrusting int32
	armed     int32
	name      string
}

func (s *shipEntity) Update(e *downstream.Entity) {
	if e.SnapshotType() != downstream.SnapshotShip {
		panic("game: tried to update ship with non-ship snapshot")
	}
	snapshotTable := new(flatbuffers.Table)
	if !e.Snapshot(snapshotTable) {
		panic("game: could not extract ship snapshot from entity during update")
	}
	snap := new(downstream.Ship)
	snap.Init(snapshotTable.Bytes, snapshotTable.Pos)
	posn := snap.Position(nil)
	//log.Print("posn X ", posn.X(), " Y ", posn.Y())
	s.physics.SetPosition(cp.Vector{X: float64(posn.X()), Y: float64(posn.Y())})
	s.physics.SetAngle(float64(snap.Rotation()))
	atomic.StoreInt32(&s.thrusting, int32(snap.Thrusting()))
	atomic.StoreInt32(&s.armed, int32(snap.Armed()))
	//if snap.Thrusting() > 0 {
	//	s.physics.ApplyForceAtLocalPoint(cp.Vector{Y: 200}, cp.Vector{}, )
	//}
	//log.Print(e.Id(), s.physics.Position(), s.physics.Angle())
}

func (s *shipEntity) Position() pixel.Vec {
	return pixel.Vec(s.physics.Position())
}

func (s *shipEntity) Draw(imd *imdraw.IMDraw) {
	imd.Color = color.RGBA{
		R: 242,
		G: 75,
		B: 105,
		A: 255,
	}
	posn := s.physics.Position()
	r := s.physics.Rotation()
	imd.Push(
		v(ship.Vertices[0].Rotate(r).Add(posn)),
		v(ship.Vertices[1].Rotate(r).Add(posn)),
		v(ship.Vertices[2].Rotate(r).Add(posn)),
	)
	imd.Polygon(0)
	if atomic.LoadInt32(&s.thrusting) > 0 {
		imd.Color = color.RGBA{
			R: 235,
			G: 200,
			B: 82,
			A: 255,
		}
		imd.Push(
			pixel.Vec{-8, -20}.Rotated(r.ToAngle()).Add(v(posn)),
			pixel.Vec{8, -20}.Rotated(r.ToAngle()).Add(v(posn)),
			pixel.Vec{0, -40}.Rotated(r.ToAngle()).Add(v(posn)),
		)
		imd.Polygon(0)
	}
	if atomic.LoadInt32(&s.armed) > 0 {
		imd.Color = color.RGBA{
			R: 74,
			G: 136,
			B: 212,
			A: 255,
		}
		imd.Push(pixel.ZV.Add(v(posn)))
		imd.Circle(8, 0)
	}
}

func v(vector cp.Vector) pixel.Vec {
	return pixel.Vec(vector)
}
