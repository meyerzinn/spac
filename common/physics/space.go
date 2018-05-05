package physics

import "github.com/jakecoffman/cp"

// NewSpace returns a new, pre-configured Space.
func NewSpace() *cp.Space {
	space := cp.NewSpace()
	// > rules
	space.SetDamping(0.3)
	space.SetGravity(cp.Vector{0, 0})
	return space
}
