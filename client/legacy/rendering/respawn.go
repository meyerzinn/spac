package rendering

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/20zinnm/spac/client/fonts"
	"github.com/faiface/pixel"
	"golang.org/x/image/colornames"
	"github.com/20zinnm/entity"
)

var (
	atlas = text.NewAtlas(fonts.RobotoLight, text.ASCII)
)

type RespawnHandler func(name string)

type RespawnScene struct {
	handler RespawnHandler

	name string
	txt  *text.Text
}

func NewRespawnScene(handler RespawnHandler) *RespawnScene {
	txt := text.New(pixel.V(50, 500), atlas)
	txt.Color = colornames.Black
	return &RespawnScene{
		handler: handler,
		txt:     txt,
	}
}

func (s *RespawnScene) Render(window *pixelgl.Window) {

}

func (s *RespawnScene) Update(delta float64) {
	panic("implement me")
}

func (s *RespawnScene) Remove(entity entity.ID) {
	panic("implement me")
}