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
	manager        *entity.Manager
	entities       map[entity.ID]Entity
	world          *physics.World
	canvas         *pixelgl.Canvas
	imd            *imdraw.IMDraw
	camPos         pixel.Vec
	camera         Camera // can be nil!
	target         entity.ID
	handler        Handler
	lastPerception int64
}

func New(win *pixelgl.Window, worldRadius float64, me entity.ID, handler Handler) *Scene {
	manager := new(entity.Manager)
	world := &physics.World{
		Space: physics.NewSpace(),
	}
	manager.AddSystem(physics.New(manager, world, worldRadius))
	return &Scene{
		win:      win,
		manager:  manager,
		world:    world,
		entities: make(map[entity.ID]Entity),
		//canvas:   pixelgl.NewCanvas(pixel.R(-1024/2, -768/2, 1024/2, 768/2)),
		canvas:  pixelgl.NewCanvas(pixel.R(-1024/2, -768/2, 1024/2, 768/2)),
		imd:     imdraw.New(nil),
		camPos:  pixel.ZV,
		target:  me,
		handler: handler,
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
			newShip.physics = ship.NewPhysics(s.world.Space, e.Id(), cp.Vector{})
			newShip.Update(e)
			s.entities[e.Id()] = newShip
			for _, system := range s.manager.Systems() {
				switch sys := system.(type) {
				case *physics.System:
					sys.Add(e.Id(), newShip.physics)
				}
			}
			if s.camera == nil && e.Id() == s.target {
				log.Print("tracking ", e.Id())
				s.camera = newShip
			}
		}
	}
	for id := range s.entities {
		if _, ok := known[id]; !ok {
			go s.manager.Remove(id)
			delete(s.entities, id)
		}
	}
}

func (s *Scene) Update(dt float64) {
	physicsDelta := time.Now().Sub(time.Unix(0, atomic.LoadInt64(&s.lastPerception))).Seconds()
	s.manager.Update(physicsDelta)
	if s.camera == nil {
		return
	}
	s.world.Lock()
	defer s.world.Unlock()
	s.camPos = pixel.Lerp(s.camPos, s.camera.Position(), 1-math.Pow(1.0/128, dt))
	//log.Print("camPos ", s.camPos, " camera position ", s.camera.Position(), " step ", 1-math.Pow(1.0/128, dt), " dt ", dt)
	cam := pixel.IM.Moved(s.camPos.Scaled(-1))
	s.canvas.SetMatrix(cam)
	s.canvas.Clear(colornames.White)
	s.imd.Clear()
	for _, renderable := range s.entities {
		renderable.Draw(s.imd)
	}
	s.imd.Draw(s.canvas)

	s.win.Clear(colornames.White)
	s.win.SetMatrix(pixel.IM.Scaled(pixel.ZV,
		math.Min(
			s.win.Bounds().W()/s.canvas.Bounds().W(),
			s.win.Bounds().H()/s.canvas.Bounds().H(),
		),
	).Moved(s.win.Bounds().Center()))
	s.canvas.Draw(s.win, pixel.IM.Moved(s.canvas.Bounds().Center()))

	s.handler.Handle(Controls{
		Left:   s.win.Pressed(pixelgl.KeyA),
		Right:  s.win.Pressed(pixelgl.KeyD),
		Thrust: s.win.Pressed(pixelgl.KeyW),
		Shoot:  s.win.Pressed(pixelgl.MouseButton1) || s.win.Pressed(pixelgl.KeySpace),
	})
}

type shipEntity struct {
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
	s.physics.SetPosition(cp.Vector{X: float64(posn.X()), Y: float64(posn.Y())})
	s.physics.SetAngle(float64(snap.Rotation()))
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
	posn := pixel.Vec(s.physics.Position())
	a := s.physics.Angle()
	imd.Push(
		pixel.Vec{-24, -20}.Rotated(a).Add(posn),
		pixel.Vec{24, -20}.Rotated(a).Add(posn),
		pixel.Vec{0, 40}.Rotated(a).Add(posn),
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
			pixel.Vec{-8, -20}.Rotated(a).Add(posn),
			pixel.Vec{8, -20}.Rotated(a).Add(posn),
			pixel.Vec{0, -40}.Rotated(a).Add(posn),
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
		imd.Push(pixel.ZV.Add(posn))
		imd.Circle(8, 0)
	}
}
