package perceiving

import (
	"fmt"
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/builders"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/server/physics/collision"
	"github.com/google/flatbuffers/go"
	"github.com/jakecoffman/cp"
	"io"
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
	space        *cp.Space
	perceivers   map[entity.ID]*perceivingEntity
	perceivables map[entity.ID]Perceivable
}

func New(space *cp.Space) *System {
	return &System{
		space:        space,
		perceivers:   make(map[entity.ID]*perceivingEntity),
		perceivables: make(map[entity.ID]Perceivable),
	}
}

func (s *System) AddPerceiver(id entity.ID, perceiver Perceiver) {
	s.perceivers[id] = &perceivingEntity{Perceiver: perceiver, known: make(map[entity.ID]struct{})}
}

func (s *System) AddPerceivable(id entity.ID, perceivable Perceivable) {
	s.perceivables[id] = perceivable
}

func (s *System) Update(_ float64) {
	for id, perceiver := range s.perceivers {
		s.perceive(id, perceiver)
	}
}

type perceivable struct {
	Perceivable
	entity.ID
}

func (s *System) perceive(id entity.ID, perceiver *perceivingEntity) {
	b := builders.Get()
	defer builders.Put(b)
	nearby := make(map[entity.ID]struct{})
	s.space.BBQuery(cp.NewBBForCircle(perceiver.Position(), 1000),
		cp.NewShapeFilter(0, uint(collision.Perceiving), uint(collision.Perceivable)),
		func(shape *cp.Shape, _ interface{}) {
			nearby[shape.Body().UserData.(entity.ID)] = struct{}{}
		},
		nil)
	perceivables := make([]perceivable, 0, len(nearby)+1)
	// include self
	if _, ok := s.perceivables[id]; ok {
		nearby[id] = struct{}{}
	}
	for eid := range nearby {
		if p, ok := s.perceivables[eid]; ok {
			perceivables = append(perceivables, perceivable{Perceivable: p, ID: eid})
		}
	}
	var entities []flatbuffers.UOffsetT
	for _, p := range perceivables {
		_, known := perceiver.known[p.ID]
		entities = append(entities, p.Snapshot(b, known))
		if !known {
			perceiver.known[p.ID] = struct{}{}
		}
	}
	downstream.PerceptionStartEntitiesVector(b, len(entities))
	for _, entity := range entities {
		b.PrependUOffsetT(entity)
	}
	entitiesVec := b.EndVector(len(entities))
	downstream.PerceptionStart(b)
	downstream.PerceptionAddEntities(b, entitiesVec)
	go perceiver.Perceive(net.MessageDown(b, downstream.PacketPerception, downstream.PerceptionEnd(b)))
}

func (s *System) Debug(w io.Writer) {
	fmt.Fprintln(w, "perceiving system")
	fmt.Fprintf(w, "perceiversCount=%d\n", len(s.perceivers))
	fmt.Fprintf(w, "perceivablesCount=%d\n", len(s.perceivables))
}

func (s *System) Remove(entity entity.ID) {
	delete(s.perceivers, entity)
	delete(s.perceivables, entity)
}
