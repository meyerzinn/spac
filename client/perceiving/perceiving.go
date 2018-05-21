package perceiving

import (
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/entity"
	"sync"
	"github.com/20zinnm/spac/client/physics"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/client/rendering"
	"github.com/faiface/pixel/imdraw"
	"image/color"
	"github.com/faiface/pixel"
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/common/world"
)

type Updater interface {
	Update(*downstream.Entity)
}

type UpdaterFunc func(*downstream.Entity)

func (fn UpdaterFunc) Update(e *downstream.Entity) {
	fn(e)
}

type System struct {
	manager    *entity.Manager
	world      *world.World
	self       entity.ID
	entitiesMu sync.RWMutex
	entities   map[entity.ID]Updater
}

func New(manager *entity.Manager, world *world.World, self entity.ID) *System {
	return &System{
		manager:  manager,
		world:    world,
		self:     self,
		entities: make(map[entity.ID]Updater),
	}
}

func (s *System) Update(delta float64) {
}

func (s *System) Perceive(perception *downstream.Perception) {
	s.entitiesMu.Lock()
	s.world.Lock()
	defer s.world.Unlock()
	defer s.entitiesMu.Unlock()

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
			id := e.Id()
			shipPhysics := physics.NewShip(s.world.Space, id)
			shipMu := new(sync.Mutex)
			var armed, thrusting bool
			for _, system := range s.manager.Systems() {
				switch sys := system.(type) {
				case *physics.System:
					sys.Add(id, shipPhysics)
				case *rendering.System:
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
							return pixel.Vec(shipPhysics.Position())
						}))
					}
				}
			}
			updater := UpdaterFunc(func(entity *downstream.Entity) {
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
				vel := snap.Velocity(nil)
				shipMu.Lock()
				defer shipMu.Unlock()
				shipPhysics.SetPosition(cp.Vector{X: float64(posn.X()), Y: float64(posn.Y())})
				shipPhysics.SetVelocity(float64(vel.X()), float64(vel.Y()))
				shipPhysics.SetAngle(float64(snap.Angle()))
				thrusting = snap.Thrusting() > 0
				armed = snap.Armed() > 0
			})
			updater.Update(e)
			s.entities[id] = updater
		}
	}
	for id := range s.entities {
		if _, ok := known[id]; !ok {
			go s.manager.Remove(id)
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	delete(s.entities, entity)
	s.entitiesMu.Unlock()
}
