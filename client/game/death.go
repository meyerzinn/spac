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
	"github.com/faiface/pixel/imdraw"
)

type DyingScene struct {
	win  *pixelgl.Window
	conn net.Connection
	text *text.Text
	imd  *imdraw.IMDraw
}

func NewDeath(win *pixelgl.Window, conn net.Connection) *DyingScene {
	win.SetMatrix(pixel.IM)
	return &DyingScene{
		win:  win,
		imd:  imdraw.New(nil),
		conn: conn,
		text: text.New(pixel.V(0, 0), fonts.Atlas),
	}
}

func (s *DyingScene) Update(dt float64) {
	if s.win.Pressed(pixelgl.KeyEnter) {
		CurrentScene = NewSpawnMenu(s.win, s.conn)
		return
	}
	s.win.Clear(colornames.Black)
	s.imd.Clear()
	stars.Static(s.imd, s.win.Bounds())
	s.imd.Draw(s.win)
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
