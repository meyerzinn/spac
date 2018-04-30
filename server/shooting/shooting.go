package shooting

import (
	"github.com/20zinnm/spac/common/physics"
	"github.com/20zinnm/entity"
	"sync"
	"github.com/20zinnm/spac/common/physics/world"
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/server/networking/fbs"
	"github.com/20zinnm/spac/server/despawning"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/server/perceiving"
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
			s.manager.NewEntity(func(id entity.ID, systems []entity.System) {
				bullet.ID = id
				for _, system := range systems {
					switch sys := system.(type) {
					case *physics.System:
						s.world.Do(func(space *cp.Space) {
							bullet.Physics.Body = space.AddBody(cp.NewBody(1, cp.MomentForCircle(1, 0, 12, cp.Vector{})))
							bulletShape := space.AddShape(cp.NewCircle(bullet.Physics.Body, 12, cp.Vector{}))
							bulletShape.SetFilter(cp.NewShapeFilter(uint(owner), cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))
							bullet.Physics.SetAngle(e.physics.Angle())
							bullet.Physics.SetVelocityVector(bullet.Physics.Rotation().Rotate(cp.Vector{0, e.BulletVelocity}))
						})
						sys.Add(id, bullet.Physics)
					case *despawning.System:
						sys.Add(id, 120)
					case *perceiving.System:
						sys.AddPerceivable(id, &bullet)
					}
				}
			})
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
	fbs.BulletStart(builder)
	position := b.Physics.Position()
	fbs.BulletAddPosition(builder, fbs.CreatePoint(builder, int32(position.X), int32(position.Y)))
	velocity := b.Physics.Velocity()
	fbs.BulletAddVelocity(builder, fbs.CreateVector(builder, float32(velocity.X), float32(velocity.Y)))
	return fbs.BulletEnd(builder)
}