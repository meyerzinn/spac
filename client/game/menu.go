package game

import (
	"golang.org/x/image/colornames"
	"github.com/faiface/pixel/pixelgl"
	"github.com/20zinnm/spac/common/net"
	"fmt"
	"github.com/faiface/pixel/text"
	"github.com/faiface/pixel"
)

type MenuScene struct {
	win    *pixelgl.Window
	conn   net.Connection
	text   *text.Text
	bounds pixel.Rect
	name   string
}

func newMenu(win *pixelgl.Window, conn net.Connection) *MenuScene {
	return &MenuScene{
		win:  win,
		conn: conn,
		//ctx:  ctx,
		text:   text.New(pixel.V(0, 0), atlas),
		bounds: win.Bounds(),
	}
}

func (s *MenuScene) Update(_ float64) {
	s.win.Clear(colornames.Black)
	s.name += s.win.Typed()
	if len(s.win.Typed()) > 0 {
		if s.win.JustPressed(pixelgl.KeyBackspace) {
			s.name = s.name[:len(s.name)-1]
		}
	}
	s.text.Clear()
	s.text.WriteString("This is the tale of: " + s.name)
	s.text.Draw(s.win, pixel.IM.Moved(s.win.Bounds().Max.Scaled(.5).Sub(s.text.Bounds().Max.Scaled(.5))))
	if s.win.JustPressed(pixelgl.KeyEnter) {
		CurrentScene = newSpawning(s.win, s.conn, s.name)
		fmt.Println("next scene (old:menu)")
	}
}
