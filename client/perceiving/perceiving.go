package perceiving

import (
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/entity"
	"sync"
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/spac/client/rendering"
	"github.com/faiface/pixel"
	"github.com/google/flatbuffers/go"
	"github.com/jakecoffman/cp"
	"github.com/faiface/pixel/imdraw"
	"image/color"
	"github.com/20zinnm/spac/client/physics"
	"time"
)

type Updater interface {
	Update(*downstream.Entity)
}

type UpdaterFunc func(*downstream.Entity)

func (fn UpdaterFunc) Update(e *downstream.Entity) {
	fn(e)
}

type System struct {
	manager *entity.Manager
	world   *world.World
	// stateMu guards self, entities, and last
	stateMu  sync.RWMutex
	self     entity.ID
	//latency  *int64
	entities map[entity.ID]Updater
	last     time.Time // ns since last update
}

func New(manager *entity.Manager, world *world.World, self entity.ID,/* latency *int64*/) *System {
	return &System{
		manager:  manager,
		world:    world,
		self:     self,
		//latency:  latency,
		entities: make(map[entity.ID]Updater),
	}
}

func (s *System) Update(delta float64) {
}

func (s *System) Perceive(perception *downstream.Perception) {
	s.stateMu.Lock()
	s.world.Lock()
	defer s.world.Unlock()
	defer s.stateMu.Unlock()

	//now := time.Now()
	//delta := now.Sub(s.last).Seconds()
	//s.last = now

	known := make(map[entity.ID]struct{}, perception.EntitiesLength())
	for i := 0; i < perception.EntitiesLength(); i++ {
		e := new(downstream.Entity)
		if !perception.Entities(e, i) {
			panic("failed to decode entity from perception vector")
			continue
		}
		known[e.Id()] = struct{}{}
		updater, ok := s.entities[e.Id()]
		if ok {
			updater.Update(e)
		} else {
			switch e.SnapshotType() {
			case downstream.SnapshotShip:
				id := e.Id()
				shipPhysics := physics.NewShip(s.world.Space, id)
				shipMu := new(sync.Mutex)
				var armed, thrusting bool
				for _, system := range s.manager.Systems() {
					switch sys := system.(type) {
					case *physics.System:
						sys.Add(id, shipPhysics)
					case *rendering.System:
						//var lastPosn pixel.Vec
						sys.Add(id, rendering.RenderableFunc(func(imd *imdraw.IMDraw) {
							imd.Color = color.RGBA{
								R: 242,
								G: 75,
								B: 105,
								A: 255,
							}
							shipMu.Lock()
							defer shipMu.Unlock()
							a := shipPhysics.Angle()
							//p := pixel.Lerp(pixel.Vec(shipPhysics.Position()), lastPosn, 1-math.Pow(1/128, delta))
							//lastPosn = pixel.Vec(shipPhysics.Position())
							p := pixel.Vec(shipPhysics.Position())
							imd.Push(
								pixel.Vec{-24, -20}.Rotated(a).Add(p),
								pixel.Vec{24, -20}.Rotated(a).Add(p),
								pixel.Vec{0, 40}.Rotated(a).Add(p),
							)
							imd.Polygon(0)
							if thrusting {
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
							if armed {
								imd.Color = color.RGBA{
									R: 74,
									G: 136,
									B: 212,
									A: 255,
								}
								imd.Push(p)
								imd.Circle(8, 0)
							}
						}))
						if id == s.self {
							sys.Track(rendering.TrackableFunc(func() pixel.Vec {
								shipMu.Lock()
								defer shipMu.Unlock()
								return pixel.Vec(shipPhysics.Position())
							}))
						}
					}
				}
				updater := UpdaterFunc(func(entity *downstream.Entity) {
					if entity.SnapshotType() != downstream.SnapshotShip {
						panic("game: tried to update ship with non-ship snapshot")
					}
					snapshotTable := new(flatbuffers.Table)
					if !entity.Snapshot(snapshotTable) {
						panic("game: could not extract ship snapshot from entity during update")
					}
					shipUpdate := new(downstream.Ship)
					shipUpdate.Init(snapshotTable.Bytes, snapshotTable.Pos)
					posn := shipUpdate.Position(new(downstream.Vector))
					vel := shipUpdate.Velocity(new(downstream.Vector))
					shipMu.Lock()
					defer shipMu.Unlock()
					shipPhysics.SetPosition(cp.Vector{X: float64(posn.X()), Y: float64(posn.Y())})
					shipPhysics.SetVelocity(float64(vel.X()), float64(vel.Y()))
					shipPhysics.SetAngle(float64(shipUpdate.Angle()))
					thrusting = shipUpdate.Thrusting() > 0
					armed = shipUpdate.Armed() > 0
				})
				updater.Update(e)
				s.entities[id] = updater
			}
		}
	}
	for id := range s.entities {
		if _, ok := known[id]; !ok {
			go s.manager.Remove(id)
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	s.stateMu.Lock()
	delete(s.entities, entity)
	s.stateMu.Unlock()
}
