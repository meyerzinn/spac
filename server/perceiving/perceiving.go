package perceiving

import (
	"sync"
	"github.com/jakecoffman/cp"
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/common/net/builders"
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/common/physics"
)

const CollisionType cp.CollisionType = 1 << 2

type Perceiver interface {
	Position() cp.Vector
	Perceive(perception []byte)
}

type Perceivable interface {
	Snapshot(builder *flatbuffers.Builder, known bool) flatbuffers.UOffsetT
}

type perceivingEntity struct {
	Perceiver
	known map[entity.ID]struct{}
}

type System struct {
	world          *physics.World
	perceiversMu   sync.RWMutex
	perceivers     map[entity.ID]perceivingEntity
	perceivablesMu sync.RWMutex
	perceivables   map[entity.ID]Perceivable
}

func New(world *physics.World) *System {
	return &System{
		world:        world,
		perceivers:   make(map[entity.ID]perceivingEntity),
		perceivables: make(map[entity.ID]Perceivable),
	}
}

func (s *System) AddPerceiver(id entity.ID, perceiver Perceiver) {
	s.perceiversMu.Lock()
	s.perceivers[id] = perceivingEntity{Perceiver: perceiver, known: make(map[entity.ID]struct{})}
	s.perceiversMu.Unlock()
}

func (s *System) AddPerceivable(id entity.ID, perceivable Perceivable) {
	s.perceivablesMu.Lock()
	s.perceivables[id] = perceivable
	s.perceivablesMu.Unlock()
}

func (s *System) Update(_ float64) {
	s.perceiversMu.RLock()
	defer s.perceiversMu.RUnlock()
	var wg sync.WaitGroup
	for id, perceiver := range s.perceivers {
		wg.Add(1)
		go s.perceive(id, perceiver, &wg)
	}
	wg.Wait()
}

func (s *System) perceive(id entity.ID, perceiver perceivingEntity, wg *sync.WaitGroup) {
	defer wg.Done()
	b := builders.Get()
	defer builders.Put(b)

	nearby := make(map[entity.ID]struct{})
	s.world.Lock()
	s.world.Space.BBQuery(cp.NewBBForCircle(perceiver.Position(), 1000), cp.NewShapeFilter(0, uint(CollisionType), uint(CollisionType)), func(shape *cp.Shape, _ interface{}) {
		nearby[shape.Body().UserData.(entity.ID)] = struct{}{}
	}, nil)
	s.world.Unlock()
	perceivables := make([]Perceivable, 0, len(nearby))
	s.perceivablesMu.RLock()
	for eid := range nearby {
		if p, ok := s.perceivables[eid]; ok {
			perceivables = append(perceivables, p)
		}
	}
	s.perceivablesMu.RUnlock()
	var snapshots []flatbuffers.UOffsetT
	for _, perceivable := range perceivables {
		_, known := perceiver.known[id]
		snapshots = append(snapshots, perceivable.Snapshot(b, known))
		if !known {
			perceiver.known[id] = struct{}{}
		}
	}
	downstream.PerceptionStartEntitiesVector(b, len(snapshots))
	for _, snapshot := range snapshots {
		b.PrependUOffsetT(snapshot)
	}
	entities := b.EndVector(len(snapshots))
	downstream.PerceptionStart(b)
	downstream.PerceptionAddEntities(b, entities)
	go perceiver.Perceive(net.MessageDown(b, downstream.PacketPerception, downstream.PerceptionEnd(b)))
}

func (s *System) Remove(entity entity.ID) {
	s.perceiversMu.Lock()
	delete(s.perceivers, entity)
	delete(s.perceivables, entity)
	s.perceiversMu.Unlock()
}
