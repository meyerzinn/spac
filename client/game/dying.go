package game

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/faiface/pixel"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/client/fonts"
)

type DyingScene struct {
	win  *pixelgl.Window
	conn net.Connection
	text *text.Text
}

func newDying(win *pixelgl.Window, conn net.Connection) *DyingScene {
	return &DyingScene{
		win:  win,
		conn: conn,
		text: text.New(pixel.V(0, 0), fonts.Atlas),
	}
}

func (s *DyingScene) Update(dt float64) {
	if s.win.Pressed(^pixelgl.Button(0)) {
		CurrentScene = newMenu(s.win, s.conn)
	}

	s.text.Draw(s.win, pixel.IM.Moved(s.win.Bounds().Max.Scaled(.5).Sub(s.text.Bounds().Max.Scaled(.5))))
}
