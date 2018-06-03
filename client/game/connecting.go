package game

import (
	"fmt"
	"github.com/20zinnm/spac/client/fonts"
	"github.com/20zinnm/spac/client/stars"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/google/flatbuffers/go"
	"github.com/gorilla/websocket"
	"golang.org/x/image/colornames"
	"log"
	"net/url"
	"github.com/faiface/pixel/imdraw"
)

type ConnectingScene struct {
	next chan Scene
	dots int
	txt  *text.Text
	win  *pixelgl.Window
	imd  *imdraw.IMDraw
}

func (s *ConnectingScene) Update(dt float64) {
	select {
	case scene := <-s.next:
		fmt.Println("next scene (old:connecting)")
		CurrentScene = scene
	default:
		s.win.Clear(colornames.Black)
		s.imd.Clear()
		stars.Static(s.imd, s.win.Bounds())
		s.imd.Draw(s.win)
		s.txt.Clear()
		s.dots++
		l := "Connecting"
		if s.dots > 120 {
			s.dots = 0
		} else if s.dots > 100 {
			l += "."
		} else if s.dots > 80 {
			l += ".."
		} else if s.dots > 60 {
			l += "..."
		} else if s.dots > 40 {
			l += ".."
		} else if s.dots > 20 {
			l += "."
		}
		s.txt.Dot.X -= s.txt.BoundsOf(l).W() / 2
		fmt.Fprintf(s.txt, l)
		s.txt.Draw(s.win, pixel.IM.Moved(s.win.Bounds().Max.Scaled(.5)).Scaled(s.win.Bounds().Max.Scaled(.5), 2))
	}
}

func newConnecting(win *pixelgl.Window, host string) *ConnectingScene {
	scene := &ConnectingScene{
		next: make(chan Scene),
		txt:  text.New(pixel.ZV, fonts.Atlas),
		win:  win,
		imd:  imdraw.New(nil),
	}
	go func() {
		u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}
		log.Printf("connecting to %s", u.String())
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		log.Print("connected")
		conn := net.Websocket(c)
		for {
			message, err := readMessage(conn)
			if err != nil {
				log.Fatalln(err)
			}
			if message.PacketType() != downstream.PacketServerSettings {
				log.Fatalln("received non-settings packet first; aborting")
			}
			packetTable := new(flatbuffers.Table)
			if !message.Packet(packetTable) {
				log.Fatalln("failed to decode settings packet")
			}
			settings := new(downstream.ServerSettings)
			settings.Init(packetTable.Bytes, packetTable.Pos)
			scene.next <- NewSpawnMenu(win, conn)
			return
		}
	}()
	return scene
}
