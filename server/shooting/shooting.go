package shooting

import (
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/entity"
	"sync"
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/server/despawning"
	"github.com/20zinnm/spac/server/perceiving"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/server/physics"
)

type shootingEntity struct {
	physics    physics.Component
	controller Controller
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

func (s *System) Add(id entity.ID, controller Controller, body physics.Component, component *Component) {
	s.entitiesMu.Lock()
	s.entities[id] = &shootingEntity{physics: body, controller: controller, Component: component}
	s.entitiesMu.Unlock()
}

func (s *System) Update(delta float64) {
	s.entitiesMu.RLock()
	defer s.entitiesMu.RUnlock()

	for owner, e := range s.entities {
		if e.Component.Armed() && e.controller.Controls().Shooting {
			var bullet bullet
			bullet.Owner = owner
			bullet.ID = s.manager.NewEntity()
			s.world.Lock()
			bullet.Physics = NewBullet(s.world.Space, bullet.ID, owner, e.physics.Angle(), e.BulletVelocity)
			s.world.Unlock()
			for _, system := range s.manager.Systems() {
				switch sys := system.(type) {
				case *physics.System:
					sys.Add(bullet.ID, bullet.Physics)
				case *despawning.System:
					sys.Add(bullet.ID, 120)
				case *perceiving.System:
					sys.AddPerceivable(bullet.ID, &bullet)
				}
			}
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	delete(s.entities, entity)
	s.entitiesMu.Unlock()
}

type bullet struct {
	ID      entity.ID
	Owner   entity.ID
	Physics physics.Component
}

func (b *bullet) Snapshot(builder *flatbuffers.Builder, known bool) flatbuffers.UOffsetT {
	downstream.BulletStart(builder)
	position := b.Physics.Position()
	downstream.BulletAddPosition(builder, downstream.CreateVector(builder, float32(position.X), float32(position.Y)))
	velocity := b.Physics.Velocity()
	downstream.BulletAddVelocity(builder, downstream.CreateVector(builder, float32(velocity.X), float32(velocity.Y)))
	snap := downstream.BulletEnd(builder)
	downstream.EntityStart(builder)
	downstream.EntityAddId(builder, b.ID)
	downstream.EntityAddSnapshot(builder, snap)
	downstream.EntityAddSnapshotType(builder, downstream.SnapshotBullet)
	return downstream.EntityEnd(builder)
}
