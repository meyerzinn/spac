package shooting

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/server/health"
	"github.com/20zinnm/spac/server/despawning"
	"github.com/20zinnm/spac/server/entities/bullet"
	"github.com/20zinnm/spac/server/perceiving"
	"github.com/20zinnm/spac/server/physics"
	"github.com/jakecoffman/cp"
	"sync"
)

type shootingEntity struct {
	controller Controller
	last       Controls
	physics    *cp.Body
	*Component
}

type System struct {
	manager    *entity.Manager
	space      *cp.Space
	entitiesMu sync.RWMutex
	entities   map[entity.ID]*shootingEntity
}

func New(manager *entity.Manager, space *cp.Space) *System {
	return &System{
		manager:  manager,
		space:    space,
		entities: make(map[entity.ID]*shootingEntity),
	}
}

func (s *System) Add(id entity.ID, component *Component, controller Controller, body *cp.Body) {
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
			physicsC := bullet.Physics(s.space, id, owner, e.physics, e.BulletForce)
			for _, system := range s.manager.Systems() {
				switch sys := system.(type) {
				case *health.System:
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
