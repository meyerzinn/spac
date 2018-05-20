package world

import (
	"github.com/jakecoffman/cp"
	"sync"
)

var WorldRadius float64 = 10000

// World wraps a cp.Space and provides synchronous access.
type World struct {
	Space *cp.Space
	sync.Mutex
}
