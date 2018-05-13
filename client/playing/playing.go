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
	"github.com/20zinnm/spac/common/physics/world"
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
	win      *pixelgl.Window
	manager  *entity.Manager
	entities map[entity.ID]Entity
	world    *world.World
	canvas   *pixelgl.Canvas
	imd      *imdraw.IMDraw
	camPos   pixel.Vec
	camera   Camera // can be nil!
	target   entity.ID
	handler  Handler
}

func New(win *pixelgl.Window, worldRadius float64, me entity.ID, handler Handler) *Scene {
	manager := new(entity.Manager)
	world := world.New(physics.NewSpace())
	manager.AddSystem(physics.New(manager, world, worldRadius))
	return &Scene{
		win:      win,
		manager:  manager,
		world:    world,
		entities: make(map[entity.ID]Entity),
		canvas:   pixelgl.NewCanvas(pixel.R(-1024/2, -768/2, 1024/2, 768/2)),
		imd:      imdraw.New(nil),
		camPos:   pixel.ZV,
		target:   me,
		handler:  handler,
	}
}

func (s *Scene) Perceive(perception *downstream.Perception) {
	known := make(map[entity.ID]struct{}, perception.EntitiesLength())
	for i := 0; i < perception.EntitiesLength(); i++ {
		e := new(downstream.Entity)
		if !perception.Entities(e, i) {
			continue
		}
		updater, ok := s.entities[e.Id()]
		if ok {
			updater.Update(e)
			continue
		}
		known[e.Id()] = struct{}{}
		switch e.SnapshotType() {
		case downstream.SnapshotShip:
			ship := new(shipEntity)
			s.world.Do(func(space *cp.Space) {
				ship.physics = physics.NewShip(space, e.Id())
			})
			ship.Update(e)
			s.entities[e.Id()] = ship
			for _, system := range s.manager.Systems() {
				switch sys := system.(type) {
				case *physics.System:
					sys.Add(e.Id(), ship.physics)
				}
			}
			if s.camera == nil && e.Id() == s.target {
				s.camera = ship
			}
		}
	}
	for id := range s.entities {
		if _, ok := known[id]; !ok {
			s.manager.Remove(id)
		}
	}
}

func (s *Scene) Update(dt float64) {
	s.manager.Update(dt)
	if s.camera == nil {
		return
	}
	//posn := s.camera.Position()
	//s.camPos = pixel.Lerp(s.camPos, posn, 1-math.Pow(1.0/128, dt))
	//cam := pixel.IM.Moved(s.camPos.Scaled(-1))
	//s.canvas.SetMatrix(cam)
	s.canvas.Clear(color.White)
	s.imd.Clear()

	for _, renderable := range s.entities {
		renderable.Draw(s.imd)
	}
	//s.win.SetMatrix(pixel.IM.Scaled(pixel.ZV,
	//	math.Min(
	//		s.win.Bounds().W()/s.canvas.Bounds().W(),
	//		s.win.Bounds().H()/s.canvas.Bounds().H(),
	//	),
	//).Moved(s.win.Bounds().Center()))
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
		A: 1,
	}
	posn := pixel.Vec(s.physics.Position())
	imd.Push(
		pixel.Vec{-24, -20}.Add(posn),
		pixel.Vec{24, -20}.Add(posn),
		pixel.Vec{0, 40}.Add(posn),
	)
	imd.Polygon(0)
	if atomic.LoadInt32(&s.thrusting) > 0 {
		imd.Color = color.RGBA{
			R: 235,
			G: 200,
			B: 82,
			A: 1,
		}
		imd.Push(
			pixel.Vec{-8, -20}.Add(posn),
			pixel.Vec{8, -20}.Add(posn),
			pixel.Vec{0, -40}.Add(posn),
		)
		imd.Polygon(0)
	}
	if atomic.LoadInt32(&s.armed) > 0 {
		imd.Color = color.RGBA{
			R: 74,
			G: 136,
			B: 212,
			A: 1,
		}
		imd.Push(pixel.ZV.Add(posn))
		imd.Circle(8, 0)
	}
}
