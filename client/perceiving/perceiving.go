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
	"time"
)

const (
	InterpolationBackTime = 100 * time.Millisecond
	InterpolationConstant = 0.01
	InterpolationBuffer = 10
)

type Updateable interface {
	Update(timestamp time.Time, data *flatbuffers.Table)
}

type System struct {
	manager     *entity.Manager
	physics     *physics.System
	perceptions chan *downstream.Perception
	self        entity.ID
	entitiesMu  sync.RWMutex
	entities    map[entity.ID]Updateable
	names       map[entity.ID]string
}

func New(manager *entity.Manager, phys *physics.System, self entity.ID, /* latency *int64*/) *System {
	return &System{
		manager:     manager,
		physics:     phys,
		self:        self,
		perceptions: make(chan *downstream.Perception, 8),
		entities:    make(map[entity.ID]Updateable),
		names:       make(map[entity.ID]string),
	}
}

func (s *System) Update(float64) {
	for {
		select {
		case p := <-s.perceptions:
			n := len(s.perceptions)
			if n != 0 {
				for i := 0; i < n; i++ {
					p = <-s.perceptions
				}
			}
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
	timestamp := time.Unix(0, perception.Timestamp())
	for i := 0; i < perception.EntitiesLength(); i++ {
		e := new(downstream.Entity)
		if !perception.Entities(e, i) {
			panic("failed to retrieve entity from perception vector")
		}
		snapshotTable := new(flatbuffers.Table)
		if !e.Snapshot(snapshotTable) {
			panic("could not decode snapshot from entity during update")
		}
		known[e.Id()] = struct{}{}
		updater, ok := s.entities[e.Id()]
		if ok {
			updater.Update(timestamp, snapshotTable)
		} else {
			id := e.Id()
			var updater Updateable
			switch e.SnapshotType() {
			case downstream.SnapshotBullet:
				bullet := NewBullet(id)
				for _, system := range s.manager.Systems() {
					switch sys := system.(type) {
					case *rendering.System:
						//var lastPosn pixel.Vec
						sys.Add(id, bullet)
					case *physics.System:
						sys.Add(id, bullet)
					}
				}
				updater = bullet
			case downstream.SnapshotShip:
				ship := NewShip(id)
				for _, system := range s.manager.Systems() {
					switch sys := system.(type) {
					case *rendering.System:
						sys.Add(id, ship)
						if id == s.self {
							sys.SetCamera(ship)
						}
					case *physics.System:
						sys.Add(id, ship)
					}
				}
				updater = ship
			}
			updater.Update(timestamp, snapshotTable)
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

func tocpv(vector *downstream.Vector) cp.Vector {
	return cp.Vector{
		X: float64(vector.X()),
		Y: float64(vector.Y()),
	}
}
