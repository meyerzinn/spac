package game

import (
	"golang.org/x/image/colornames"
	"github.com/faiface/pixel/pixelgl"
	"github.com/20zinnm/spac/common/net"
	"fmt"
	"github.com/faiface/pixel/text"
)

type MenuScene struct {
	win  *pixelgl.Window
	conn net.Connection
	text *text.Text
	name string
}

func newMenu(win *pixelgl.Window, conn net.Connection) *MenuScene {
	return &MenuScene{
		win:  win,
		conn: conn,
		//ctx:  ctx,
		//text: text.New(pixel.V(100, 100), nil),
	}
}

func (s *MenuScene) Update(_ float64) {
	s.win.Clear(colornames.Black)
	s.name += s.win.Typed()
	if len(s.win.Typed()) > 0 {
	}
	if s.win.JustPressed(pixelgl.KeyDelete) {
		s.name = ""
	}
	if s.win.JustPressed(pixelgl.KeyEnter) {
		CurrentScene = newSpawning(s.win, s.conn, s.name)
		fmt.Println("next scene (old:menu)")
	}
}
