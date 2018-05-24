package perceiving

import (
	"sync"
	"github.com/jakecoffman/cp"
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/common/net/builders"
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/spac/server/physics/collision"
	"github.com/20zinnm/spac/common/net"
)

type Perceiver interface {
	Position() cp.Vector
	Perceive(perception []byte)
}

type Perceivable interface {
	Snapshot(builder *flatbuffers.Builder, known bool) flatbuffers.UOffsetT
}

type PerceivableFunc func(builder *flatbuffers.Builder, known bool) flatbuffers.UOffsetT

func (fn PerceivableFunc) Snapshot(builder *flatbuffers.Builder, known bool) flatbuffers.UOffsetT {
	return fn(builder, known)
}

type perceivingEntity struct {
	Perceiver
	known map[entity.ID]struct{}
}

type System struct {
	world          *world.World
	perceiversMu   sync.RWMutex
	perceivers     map[entity.ID]perceivingEntity
	perceivablesMu sync.RWMutex
	perceivables   map[entity.ID]Perceivable
}

func New(world *world.World) *System {
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
	s.world.Space.BBQuery(cp.NewBBForCircle(perceiver.Position(), 1000),
		cp.NewShapeFilter(0, uint(collision.Perceiving), uint(collision.Perceiving)),
		func(shape *cp.Shape, _ interface{}) {
			nearby[shape.Body().UserData.(entity.ID)] = struct{}{}
		},
		nil)
	s.world.Unlock()
	perceivables := make([]struct {
		Perceivable
		entity.ID
	}, 0, len(nearby)+1)
	s.perceivablesMu.RLock()
	// include self
	if _, ok := s.perceivables[id]; ok {
		nearby[id] = struct{}{}
	}
	for eid := range nearby {
		if p, ok := s.perceivables[eid]; ok {
			perceivables = append(perceivables, struct {
				Perceivable
				entity.ID
			}{
				Perceivable: p,
				ID:          eid,
			})
		}
	}
	s.perceivablesMu.RUnlock()
	var entities []flatbuffers.UOffsetT
	s.world.RLock()
	for _, perceivable := range perceivables {
		_, known := perceiver.known[perceivable.ID]
		entities = append(entities, perceivable.Snapshot(b, known))
		if !known {
			perceiver.known[perceivable.ID] = struct{}{}
		}
	}
	s.world.RUnlock()
	downstream.PerceptionStartEntitiesVector(b, len(entities))
	for _, entity := range entities {
		b.PrependUOffsetT(entity)
	}
	entitiesVec := b.EndVector(len(entities))
	downstream.PerceptionStart(b)
	downstream.PerceptionAddEntities(b, entitiesVec)
	perception := net.MessageDown(b, downstream.PacketPerception, downstream.PerceptionEnd(b))
	//fmt.Println(perception)
	go perceiver.Perceive(perception)
}

func (s *System) Remove(entity entity.ID) {
	s.perceiversMu.Lock()
	delete(s.perceivers, entity)
	delete(s.perceivables, entity)
	s.perceiversMu.Unlock()
}
