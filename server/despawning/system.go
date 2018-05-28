package despawning

import (
	"fmt"
	"github.com/20zinnm/entity"
	"io"
)

type System struct {
	manager  *entity.Manager
	entities map[entity.ID]*Component
}

func New(manager *entity.Manager) *System {
	return &System{manager: manager, entities: make(map[entity.ID]*Component)}
}

func (s *System) Add(entity entity.ID, component Component) {
	s.entities[entity] = &component
}

func (s *System) Update(delta float64) {
	for id, component := range s.entities {
		component.alive++
		if component.alive >= component.TTL {
			go s.manager.Remove(id)
			delete(s.entities, id)
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	delete(s.entities, entity)
}

func (s *System) Debug(w io.Writer) {
	fmt.Fprintln(w, "---")
	fmt.Fprintln(w, "despawning system")
	fmt.Fprintf(w, "count=%d\n", len(s.entities))
	fmt.Fprintln(w, "entities=")
	for id, component := range s.entities {
		fmt.Fprintf(w, "> id=%d ttl=%d alive=%d\n", id, component.TTL, component.alive)
	}
}
