package rendering

import (
	"github.com/20zinnm/entity"
	"sync/atomic"
	"github.com/faiface/pixel/pixelgl"
)

type Renderer interface {
	SetScene(Scene)
}

type System struct {
	scene  atomic.Value
	window *pixelgl.Window
}

func NewRenderer(window *pixelgl.Window) *System {
	return &System{
		window: window,
	}
}

func (s *System) SetScene(scene Scene) {
	s.scene.Store(scene)
}

func (s *System) Update(delta float64) {
	val := s.scene.Load()
	if val != nil {
		scene, ok := val.(Scene)
		if ok {
			scene.Render(s.window)
			s.window.Update()
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	panic("implement me")
}
