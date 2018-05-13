package despawning

import (
	"github.com/20zinnm/entity"
	"sync"
)

type System struct {
	manager    *entity.Manager
	entitiesMu sync.RWMutex
	entities   map[entity.ID]*Component
}

func New(manager *entity.Manager) *System {
	return &System{manager: manager, entities: make(map[entity.ID]*Component)}
}

func (s *System) Add(entity entity.ID, ttl int) {
	s.entitiesMu.Lock()
	s.entities[entity] = &Component{TTL: uint(ttl), alive: 0}
	s.entitiesMu.Unlock()
}

func (s *System) Update(delta float64) {
	s.entitiesMu.RLock()
	defer s.entitiesMu.RUnlock()

	for id, component := range s.entities {
		component.alive++
		if component.alive >= component.TTL {
			go s.manager.Remove(id)
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	delete(s.entities, entity)
	s.entitiesMu.Unlock()
}
