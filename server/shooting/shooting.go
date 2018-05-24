package shooting

import (
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/entity"
	"sync"
	"github.com/20zinnm/spac/server/despawning"
	"github.com/20zinnm/spac/server/perceiving"
	"github.com/20zinnm/spac/server/physics"
	"github.com/20zinnm/spac/server/entities/bullet"
)

type shootingEntity struct {
	controller Controller
	last       Controls
	physics    world.Component
	*Component
}

type System struct {
	manager    *entity.Manager
	world      *world.World
	entitiesMu sync.RWMutex
	entities   map[entity.ID]*shootingEntity
}

func New(manager *entity.Manager, world *world.World) *System {
	return &System{
		manager:  manager,
		world:    world,
		entities: make(map[entity.ID]*shootingEntity),
	}
}

func (s *System) Add(id entity.ID, component *Component, controller Controller, body world.Component) {
	s.entitiesMu.Lock()
	s.entities[id] = &shootingEntity{physics: body, controller: controller, Component: component}
	s.entitiesMu.Unlock()
}

func (s *System) Update(delta float64) {
	s.entitiesMu.RLock()
	defer s.entitiesMu.RUnlock()

	for owner, e := range s.entities {
		select {
		case n := <-e.controller:
			e.last = n
		default:
		}
		if e.tta > 0 {
			e.tta--
		}
		if e.tta == 0 && e.last.Shooting {
			id := s.manager.NewEntity()
			s.world.Lock()
			physicsC := bullet.Physics(s.world.Space, id, owner, e.physics, e.BulletForce)
			s.world.Unlock()
			for _, system := range s.manager.Systems() {
				switch sys := system.(type) {
				case *physics.System:
					sys.Add(id, physicsC)
				case *despawning.System:
					sys.Add(id, bullet.Despawning(e.BulletLifetime))
				case *perceiving.System:
					sys.AddPerceivable(id, bullet.Perceivable(id, physicsC))
				}
			}
			e.tta = e.Cooldown
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	delete(s.entities, entity)
	s.entitiesMu.Unlock()
}
