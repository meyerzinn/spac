package physics

import (
	"github.com/jakecoffman/cp"
	"sync"
)

// World wraps a cp.Space and provides synchronous access.
type World struct {
	Space *cp.Space
	sync.Mutex
}
