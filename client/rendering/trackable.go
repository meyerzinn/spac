package rendering

import "github.com/faiface/pixel"

type Trackable interface {
	Position() pixel.Vec
}

type TrackableFunc func() pixel.Vec

func (fn TrackableFunc) Position() pixel.Vec {
	return fn()
}
