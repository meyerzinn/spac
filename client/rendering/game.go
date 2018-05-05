package rendering

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/20zinnm/entity"
	"github.com/faiface/pixel"
	"github.com/20zinnm/spac/common/physics"
)

type renderingEntity struct {
	Sprite  *pixel.Sprite
	Physics physics.Component
}

type GameRenderer struct {
	window   *pixelgl.Window
	entities map[entity.ID]renderingEntity
}

func (s *GameRenderer) Update(delta float64) {
	panic("implement me")
}

func (s *GameRenderer) Remove(entity entity.ID) {
	panic("implement me")
}

