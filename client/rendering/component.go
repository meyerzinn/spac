package rendering

import (
	"github.com/faiface/pixel"
	"github.com/20zinnm/spac/common/physics"
)

type Component struct {
	*pixel.Sprite
	Physics physics.Component
}