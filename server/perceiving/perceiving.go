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
	"sort"
	"sync"
	"time"
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
	last         time.Time
}

func New(space *cp.Space) *System {
	return &System{
		space:        space,
		perceivers:   make(map[entity.ID]*perceivingEntity),
		perceivables: make(map[entity.ID]Perceivable),
		last:         time.Now(),
	}
}

func (s *System) AddPerceiver(id entity.ID, perceiver Perceiver) {
	s.perceivers[id] = &perceivingEntity{Perceiver: perceiver, known: make(map[entity.ID]struct{})}
}

func (s *System) AddPerceivable(id entity.ID, perceivable Perceivable) {
	s.perceivables[id] = perceivable
}

func insertNearbyEntity(dst *[]entity.ID, id entity.ID) {
	i := sort.Search(len(*dst), func(i int) bool {
		return (*dst)[i] <= id
	})
	if i == len(*dst) || (*dst)[i] != id {
		*dst = append(*dst, 0)
		copy((*dst)[i+1:], (*dst)[i:])
		(*dst)[i] = id
	}
}

func (s *System) Update(_ float64) {
	now := time.Now()
	if now.Sub(s.last).Milliseconds() < 50 {
		return
	}
	s.last = now
	timestamp := now.UnixNano()
	var wg sync.WaitGroup
	for id, perceiver := range s.perceivers {
		nearby := make([]entity.ID, 0, 100)
		s.space.BBQuery(cp.NewBBForCircle(perceiver.Position(), 1000),
			cp.NewShapeFilter(0, uint(collision.Perceiving), uint(collision.Perceivable)),
			func(shape *cp.Shape, _ interface{}) {
				insertNearbyEntity(&nearby, shape.Body().UserData.(entity.ID))
			},
			nil)
		wg.Add(1)
		go s.perceive(id, nearby, perceiver, timestamp, &wg)
	}
	wg.Wait()
}

type perceivable struct {
	Perceivable
	entity.ID
}

func (s *System) perceive(id entity.ID, nearby []entity.ID, perceiver *perceivingEntity, timestamp int64, wg *sync.WaitGroup) {
	defer wg.Done()

	b := builders.Get()
	defer builders.Put(b)
	perceivables := make([]perceivable, 0, len(nearby)+1)
	// include self
	if _, ok := s.perceivables[id]; ok {
		insertNearbyEntity(&nearby, id)
	}
	for _, eid := range nearby {
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
	for _, e := range entities {
		b.PrependUOffsetT(e)
	}
	entitiesVec := b.EndVector(len(entities))
	downstream.PerceptionStart(b)
	downstream.PerceptionAddTimestamp(b, timestamp)
	downstream.PerceptionAddEntities(b, entitiesVec)
	go perceiver.Perceive(net.MessageDown(b, downstream.PacketPerception, downstream.PerceptionEnd(b)))
}

func (s *System) Debug(w io.Writer) {
	fmt.Fprintln(w, "perceiving system")
	fmt.Fprintf(w, "perceiversCount=%d\n", len(s.perceivers))
	fmt.Fprintf(w, "perceivablesCount=%d\n", len(s.perceivables))
	fmt.Fprintf(w, "last=%v", s.last.Format(time.RFC3339Nano))
}

func (s *System) Remove(entity entity.ID) {
	delete(s.perceivers, entity)
	delete(s.perceivables, entity)
}
