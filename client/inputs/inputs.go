package inputs

import (
	"github.com/20zinnm/entity"
	"github.com/faiface/pixel/pixelgl"
)

type Handler interface {
	Handle(Controls)
}

type HandlerFunc func(Controls)

func (h HandlerFunc) Handle(c Controls) {
	h(c)
}

type System struct {
	window  *pixelgl.Window
	handler Handler
	last    Controls
}

func New(window *pixelgl.Window, handler Handler) *System {
	return &System{
		window:  window,
		handler: handler,
	}
}

func (s *System) Update(delta float64) {
	if !s.window.Closed() {
		controls := Controls{}
		if s.window.Pressed(pixelgl.KeyA) {
			controls.Left = true
		}
		if s.window.Pressed(pixelgl.KeyD) {
			controls.Right = true
		}
		if s.window.Pressed(pixelgl.KeyW) {
			controls.Thrust = true
		}
		if s.window.Pressed(pixelgl.MouseButton1) {
			controls.Shoot = true
		}
		if controls != s.last {
			s.handler.Handle(controls)
		}
	}
}

func (s *System) Remove(_ entity.ID) {
	return
}

type Controls struct {
	Left, Right, Thrust, Shoot bool
}
