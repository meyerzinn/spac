package health

import (
	"sync"
	"github.com/20zinnm/entity"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/spac/server/physics/collision"
)

type System struct {
	manager    *entity.Manager
	entitiesMu sync.RWMutex
	entities   map[entity.ID]*Component
}

func New(w *world.World) *System {
	w.Lock()
	defer w.Unlock()
	var system System
	system.entities = make(map[entity.ID]*Component)
	handler := w.Space.NewCollisionHandler(collision.Health, collision.Health)
	handler.PreSolveFunc = func(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
		system.entitiesMu.RLock()
		defer system.entitiesMu.RUnlock()

		ab, bb := arb.Bodies()
		if aid, ok := ab.UserData.(entity.ID); ok {
			if bid, ok := bb.UserData.(entity.ID); ok {
				if a, ok := system.entities[aid]; ok {
					if b, ok := system.entities[bid]; ok {
						ao, bo := *a, *b
						*a -= bo
						*b -= ao
					}
				}
			}
		}
		return true
	}
	return &system

}

func (s *System) Add(entity entity.ID, component *Component) {
	s.entitiesMu.Lock()
	s.entities[entity] = component
	s.entitiesMu.Unlock()
}

func (s *System) Update(delta float64) {
	s.entitiesMu.RLock()
	defer s.entitiesMu.RUnlock()
	for id, d := range s.entities {
		if *d <= 0 {
			go s.manager.Remove(id)
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	delete(s.entities, entity)
	s.entitiesMu.Unlock()
}
