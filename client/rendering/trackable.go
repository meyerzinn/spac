package rendering

import "github.com/faiface/pixel"

type Camera interface {
	Position() pixel.Vec
	Health() int
}
