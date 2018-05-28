package collision

import "github.com/jakecoffman/cp"

const (
	Damageable cp.CollisionType = 1 << (iota + 1)
	Perceiving
	Perceivable
)
