package rendering

import "github.com/faiface/pixel"

type Trackable interface {
	Position() pixel.Vec
	Health() int
}
