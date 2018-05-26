package game

import (
	"golang.org/x/image/colornames"
	"github.com/faiface/pixel/pixelgl"
	"github.com/20zinnm/spac/common/net"
	"fmt"
	"github.com/faiface/pixel/text"
	"github.com/faiface/pixel"
	"github.com/20zinnm/spac/client/fonts"
)

type MenuScene struct {
	win     *pixelgl.Window
	conn    net.Connection
	text    *text.Text
	cursori int
	name    string
}

func newMenu(win *pixelgl.Window, conn net.Connection) *MenuScene {
	return &MenuScene{
		win:  win,
		conn: conn,
		text: text.New(pixel.V(0, 0), fonts.Atlas),
	}
}

func (s *MenuScene) Update(_ float64) {
	s.win.Clear(colornames.Black)
	s.name += s.win.Typed()
	if s.win.JustPressed(pixelgl.KeyBackspace) {
		if len(s.name) > 0 {
			s.name = s.name[:len(s.name)-1]
		}
	}
	s.cursori++
	cursor := " "
	if s.cursori >= 60 {
		s.cursori = 0
	}
	if s.cursori >= 30 {
		cursor = "_"
	}
	lines := []string{
		fmt.Sprintf("This is the tale of: %s%s", s.name, cursor),
		"(press enter to spawn)",
	}
	s.text.Clear()
	for _, l := range lines {
		s.text.Dot.X -= s.text.BoundsOf(l).W() / 2
		fmt.Fprintln(s.text, l)
	}
	s.text.Draw(s.win, pixel.IM.Moved(s.win.Bounds().Max.Scaled(.5)).Scaled(s.win.Bounds().Max.Scaled(.5), 2))
	if s.win.JustPressed(pixelgl.KeyEnter) {
		CurrentScene = newSpawning(s.win, s.conn, s.name)
		fmt.Println("next scene (old:menu)")
	}
}
