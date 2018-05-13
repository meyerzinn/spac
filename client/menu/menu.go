package menu

import (
	"github.com/faiface/pixel/pixelgl"
	"log"
	"golang.org/x/image/colornames"
)

type Handler interface {
	Spawn(name string)
}

type HandlerFunc func(name string)

func (fn HandlerFunc) Spawn(name string) {
	fn(name)
}

type Scene struct {
	win     *pixelgl.Window
	name    string
	handler Handler
}

func New(win *pixelgl.Window, handler Handler) *Scene {
	win.Clear(colornames.Blue)
	return &Scene{win: win, handler: handler,}
}

func (s *Scene) Update(dt float64) {
	s.name += s.win.Typed()
	if len(s.win.Typed()) > 0 {
		log.Print(s.win.Typed(), " ", s.name)
	}
	if s.win.JustPressed(pixelgl.KeyDelete) {
		s.name = ""
	}
	if s.win.JustPressed(pixelgl.KeyEnter) {
		s.handler.Spawn(s.name)
	}
}
