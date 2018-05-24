package perceiving

import (
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/entity"
	"sync"
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/spac/client/rendering"
	"github.com/20zinnm/spac/client/physics"
	"time"
	"github.com/20zinnm/spac/client/entities/ship"
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/client/entities/bullet"
	"github.com/faiface/pixel"
	"fmt"
)

type Updater interface {
	Update(bytes []byte, pos flatbuffers.UOffsetT)
}

type UpdaterFunc func([]byte, flatbuffers.UOffsetT)

func (fn UpdaterFunc) Update(bytes []byte, pos flatbuffers.UOffsetT) {
	fn(bytes, pos)
}

type System struct {
	manager *entity.Manager
	world   *world.World
	// stateMu guards self, entities, and last
	stateMu sync.RWMutex
	self    entity.ID
	//latency  *int64
	entities map[entity.ID]Updater
	last     time.Time // ns since last update
}

func New(manager *entity.Manager, world *world.World, self entity.ID, /* latency *int64*/) *System {
	return &System{
		manager: manager,
		world:   world,
		self:    self,
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
			panic("failed to retrieve entity from perception vector")
			continue
		}
		snapshotTable := new(flatbuffers.Table)
		if !e.Snapshot(snapshotTable) {
			panic("could not decode snapshot from entity during update")
		}
		known[e.Id()] = struct{}{}
		updater, ok := s.entities[e.Id()]
		if ok {
			updater.Update(snapshotTable.Bytes, snapshotTable.Pos)
		} else {
			id := e.Id()
			var updater Updater
			switch e.SnapshotType() {
			case downstream.SnapshotShip:
				physicsC := ship.Physics(s.world.Space, id)
				var armed, thrusting bool
				for _, system := range s.manager.Systems() {
					switch sys := system.(type) {
					case *physics.System:
						sys.Add(id, physicsC)
					case *rendering.System:
						//var lastPosn pixel.Vec
						sys.Add(id, ship.Renderable(physicsC, &thrusting, &armed))
						if id == s.self {
							sys.Track(rendering.TrackableFunc(func() pixel.Vec {
								return pixel.Vec(physicsC.Position())
							}))
						}
					}
				}
				updater = UpdaterFunc(ship.Updater(physicsC, &thrusting, &armed))
			case downstream.SnapshotBullet:
				id := e.Id()
				physicsC := bullet.Physics(s.world.Space)
				for _, system := range s.manager.Systems() {
					switch sys := system.(type) {
					case *physics.System:
						sys.Add(id, physicsC)
					case *rendering.System:
						sys.Add(id, bullet.Renderable(physicsC))
					}
				}
				updater = UpdaterFunc(bullet.Updater(physicsC))
			}
			updater.Update(snapshotTable.Bytes, snapshotTable.Pos)
			s.entities[id] = updater
		}
	}
	for id := range s.entities {
		if _, ok := known[id]; !ok {
			s.manager.Remove(id)
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	s.stateMu.Lock()
	delete(s.entities, entity)
	if entity == s.self {
		fmt.Println("died")
	}
	s.stateMu.Unlock()
}
