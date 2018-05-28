package game

import (
	"fmt"
	"github.com/20zinnm/spac/client/fonts"
	"github.com/20zinnm/spac/client/stars"
	"github.com/20zinnm/spac/common/net"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
)

type DyingScene struct {
	win  *pixelgl.Window
	conn net.Connection
	text *text.Text
}

func NewDeath(win *pixelgl.Window, conn net.Connection) *DyingScene {
	return &DyingScene{
		win:  win,
		conn: conn,
		text: text.New(pixel.V(0, 0), fonts.Atlas),
	}
}

func (s *DyingScene) Update(dt float64) {
	if s.win.Pressed(pixelgl.KeyEnter) {
		CurrentScene = NewSpawnMenu(s.win, s.conn)
		return
	}
	s.win.SetMatrix(pixel.IM)
	s.win.Clear(colornames.Black)
	stars.Static(s.win)
	lines := []string{
		"It was a good run.",
		"(press enter to continue)",
	}
	s.text.Clear()
	for _, l := range lines {
		s.text.Dot.X -= s.text.BoundsOf(l).W() / 2
		fmt.Fprintln(s.text, l)
	}
	s.text.Draw(s.win, pixel.IM.Moved(s.win.Bounds().Max.Scaled(.5)).Scaled(s.win.Bounds().Max.Scaled(.5), 2))
}
