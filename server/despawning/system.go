package despawning

import (
	"github.com/20zinnm/entity"
	"sync"
	"io"
	"fmt"
)

type System struct {
	manager    *entity.Manager
	entitiesMu sync.RWMutex
	entities   map[entity.ID]*Component
}

func New(manager *entity.Manager) *System {
	return &System{manager: manager, entities: make(map[entity.ID]*Component)}
}

func (s *System) Add(entity entity.ID, component Component) {
	s.entitiesMu.Lock()
	s.entities[entity] = &component
	s.entitiesMu.Unlock()
}

func (s *System) Update(delta float64) {
	s.entitiesMu.RLock()
	defer s.entitiesMu.RUnlock()

	for id, component := range s.entities {
		component.alive++
		if component.alive >= component.TTL {
			go s.manager.Remove(id)
			delete(s.entities, id)
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	delete(s.entities, entity)
	s.entitiesMu.Unlock()
}

func (s *System) Debug(w io.Writer) {
	s.entitiesMu.RLock()
	defer s.entitiesMu.RUnlock()
	fmt.Fprintln(w, "---")
	fmt.Fprintln(w, "despawning system")
	fmt.Fprintf(w, "count=%d\n", len(s.entities))
	fmt.Fprintln(w, "entities=")
	for id, component := range s.entities {
		fmt.Fprintf(w, "> id=%d ttl=%d alive=%d\n", id, component.TTL, component.alive)
	}
}
