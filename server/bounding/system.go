package bounding

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/server/health"
	"github.com/jakecoffman/cp"
	"math"
)

type boundingEntity struct {
	Physics *cp.Body
	Health  *health.Component
}

type System struct {
	radius   float64
	entities map[entity.ID]boundingEntity
}

func New(radius float64) *System {
	return &System{
		radius: radius,
	}
}

func (s *System) Add(id entity.ID, physics *cp.Body, health *health.Component) {
	s.entities[id] = boundingEntity{Physics: physics, Health: health,}
}

func (s *System) Update(delta float64) {
	for _, e := range s.entities {
		if !e.Physics.Position().Near(cp.Vector{}, s.radius) {
			e.Health.Value -= damage(e.Physics.Position().Length() - s.radius)
		}
	}
}

func damage(dist float64) float64 {
	return math.Sqrt(dist) / 20
}

func (s *System) Remove(entity entity.ID) {
	delete(s.entities, entity)
}

//
//type boundingEntity struct {
//	Physics world.Component
//	Health  *damaging.Component
//	Decay   float64
//}
//
//type System struct {
//	world      *world.World
//	radius     float64
//	entitiesMu sync.RWMutex
//	entities   map[entity.ID]boundingEntity
//}
//
//func New(world *world.World, radius float64) *System {
//	return &System{
//		world:    world,
//		radius:   radius,
//		entities: make(map[entity.ID]boundingEntity),
//	}
//}
//
//func (s *System) Add(id entity.ID, constants world.Component, health *damaging.Component, decay float64) {
//	s.entitiesMu.Lock()
//	s.entities[id] = boundingEntity{
//		Physics: constants,
//		Health:  health,
//		Decay:   decay,
//	}
//	s.entitiesMu.Unlock()
//}
//
//func (s *System) Update(delta float64) {
//	s.entitiesMu.RLock()
//	defer s.entitiesMu.RUnlock()
//	for _, entity := range s.entities {
//		if !entity.Physics.Position().Near(cp.Vector{}, s.radius) {
//			entity.Health.Value -= entity.Decay * entity.Health.Max
//		}
//	}
//}
//
//func (s *System) Remove(entity entity.ID) {
//	panic("implement me")
//}
