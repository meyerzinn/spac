package world

import (
	"github.com/jakecoffman/cp"
	"sync"
)

// World wraps a cp.Space and provides synchronous access.
type World struct {
	// RWMutex protects Space. Read-locking it indicates that a read-only operation is currently operating on dependent objects, such as reading a Body's position.
	sync.RWMutex
	Space *cp.Space
}
