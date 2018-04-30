package world

import (
	"github.com/jakecoffman/cp"
	"sync"
)

// World provides synchronized access to cp.Space.
type World struct {
	spaceMu sync.Mutex
	space   *cp.Space
}

func New(space *cp.Space) *World {
	return &World{
		space: space,
	}
}

// Do blocks until fn has been called. It ensures that cp.Space is not concurrently accessed.
func (w *World) Do(fn func(space *cp.Space)) {
	w.spaceMu.Lock()
	fn(w.space)
	w.spaceMu.Unlock()
}
