package perceiving

import (
	"fmt"
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/client/physics"
	"github.com/20zinnm/spac/client/rendering"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/google/flatbuffers/go"
	"github.com/jakecoffman/cp"
	"sync"
)

type Updateable interface {
	Update(*flatbuffers.Table)
}

type System struct {
	manager     *entity.Manager
	space       *cp.Space
	perceptions chan *downstream.Perception
	self        entity.ID
	entitiesMu  sync.RWMutex
	entities    map[entity.ID]Updateable
}

func New(manager *entity.Manager, space *cp.Space, self entity.ID /* latency *int64*/) *System {
	return &System{
		manager:     manager,
		space:       space,
		self:        self,
		perceptions: make(chan *downstream.Perception, 8),
		entities:    make(map[entity.ID]Updateable),
	}
}

func (s *System) Update(delta float64) {
	for {
		select {
		case p := <-s.perceptions:
			s.processPerception(p)
		default:
			return
		}
	}
}

func (s *System) Perceive(perception *downstream.Perception) {
	s.perceptions <- perception
}

func (s *System) processPerception(perception *downstream.Perception) {
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
			updater.Update(snapshotTable)
		} else {
			id := e.Id()
			var updater Updateable
			switch e.SnapshotType() {
			case downstream.SnapshotBullet:
				bullet := NewBullet(s.space, id)
				for _, system := range s.manager.Systems() {
					switch sys := system.(type) {
					case *physics.System:
						sys.Add(id, bullet.Physics)
					case *rendering.System:
						//var lastPosn pixel.Vec
						sys.Add(id, bullet)
					}
				}
				updater = bullet
			case downstream.SnapshotShip:
				ship := NewShip(s.space, id)
				for _, system := range s.manager.Systems() {
					switch sys := system.(type) {
					case *physics.System:
						sys.Add(id, ship.Physics)
					case *rendering.System:
						sys.Add(id, ship)
						if id == s.self {
							sys.Track(ship)
						}
					}
				}
				updater = ship
			}
			updater.Update(snapshotTable)
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
	delete(s.entities, entity)
	if entity == s.self {
		fmt.Println("died")
	}
}
