package world

import (
	"github.com/20zinnm/spac/common/constants"
	"github.com/jakecoffman/cp"
)

// NewSpace returns a new, pre-configured instance of cp.Space.
func NewSpace() *cp.Space {
	space := cp.NewSpace()
	// rules:
	space.SetDamping(constants.Damping)
	space.SetGravity(cp.Vector{0, 0}) // no gravity
	return space
}
