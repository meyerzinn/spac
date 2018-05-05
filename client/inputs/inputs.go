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
	//last    Controls
	//queue   chan Controls
}

func New(window *pixelgl.Window, handler Handler) *System {
	return &System{
		window:  window,
		handler: handler,
		//queue:   make(chan Controls, 1), // buffer one so rendering doesn't block on this
	}
}

//
//func (s *System) Set(c Controls) {
//	s.queue <- c
//}

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
		s.handler.Handle(controls)
	}
	//select {
	//case controls := <-s.queue:
	//	if s.last == controls {
	//		return
	//	}
	//	s.last = controls
	//	b := builders.Get()
	//	fbs.ControlsStart(b)
	//	fbs.ControlsAddLeft(b, boolToByte(controls.Left))
	//	fbs.ControlsAddRight(b, boolToByte(controls.Right))
	//	fbs.ControlsAddThrusting(b, boolToByte(controls.Thrust))
	//	fbs.ControlsAddShooting(b, boolToByte(controls.Shoot))
	//	go s.conn.Write(net.Message(b, fbs.ControlsEnd(b), fbs.PacketControls))
	//}
}

//func boolToByte(val bool) byte {
//	if val {
//		return 1
//	} else {
//		return 0
//	}
//}

func (s *System) Remove(_ entity.ID) {
	return
}

type Controls struct {
	Left, Right, Thrust, Shoot bool
}
