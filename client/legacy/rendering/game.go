package rendering

import (
	"github.com/20zinnm/entity"
	"github.com/faiface/pixel"
	"sync"
	"github.com/20zinnm/spac/common/physics"
	"github.com/faiface/pixel/pixelgl"
	"github.com/20zinnm/spac/client/inputs"
	"image/color"
)

type GameScene struct {
	manager    *entity.Manager
}

type Controller interface {
	Set(inputs.Controls)
}

type gameEntity struct {
	Sprite  *pixel.Sprite
	Physics physics.Component
}

type GameRenderer struct {
	renderer   Renderer
	manager    *entity.Manager
	controller Controller
	tracked    entity.ID

	entitiesMu sync.RWMutex
	entities   map[entity.ID]gameEntity

	camPos pixel.Vec
}

func NewGameRenderer(renderer Renderer, manager *entity.Manager, controller Controller, tracked entity.ID) *GameRenderer {
	return &GameRenderer{
		renderer:   renderer,
		manager:    manager,
		controller: controller,
		tracked:    tracked,
		entities:   make(map[entity.ID]gameEntity),
	}
}

func (s *GameRenderer) Add(entity entity.ID, sprite *pixel.Sprite, physics physics.Component) {
	s.entitiesMu.Lock()
	s.entities[entity] = gameEntity{Sprite: sprite, Physics: physics}
	s.entitiesMu.Unlock()
}

func (s *GameRenderer) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	delete(s.entities, entity)
	s.entitiesMu.Unlock()
}

func (s *GameRenderer) Update(delta float64) {
}

func (s *GameRenderer) Render(window *pixelgl.Window) {
	// update inputs
	controls := inputs.Controls{
		Left:   window.Pressed(pixelgl.KeyA),
		Right:  window.Pressed(pixelgl.KeyD),
		Thrust: window.Pressed(pixelgl.KeyW),
		Shoot:  window.Pressed(pixelgl.MouseButton1),
	}
	s.controller.Set(controls)

	// render entities
	s.entitiesMu.RLock()
	defer s.entitiesMu.RUnlock()
	self, ok := s.entities[s.tracked]
	if ok {
		s.camPos = pixel.Vec(self.Physics.Position())
	}
	cam := pixel.IM.Scaled(s.camPos, 1).Moved(window.Bounds().Center().Sub(s.camPos))
	window.Clear(color.Black)
	for _, e := range s.entities {
		posn := e.Physics.Position()
		e.Sprite.Draw(window, pixel.IM.Scaled(pixel.ZV, 1).Moved(cam.Unproject(pixel.Vec(posn))).Rotated(pixel.ZV, e.Physics.Body.Angle()))
	}
}

func (s *GameRenderer) Destroy() {
	s.manager.RemoveSystem(s)
}
