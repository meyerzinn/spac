package game

import (
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/downstream"
	"log"
	"github.com/google/flatbuffers/go"
	"net/url"
	"github.com/gorilla/websocket"
	"fmt"
	"github.com/faiface/pixel/pixelgl"
)

type ConnectingScene struct {
	next chan Scene
}

func (s *ConnectingScene) Update(dt float64) {
	select {
	case scene := <-s.next:
		fmt.Println("next scene (old:connecting)")
		CurrentScene = scene
	default:
	}
}

func newConnecting(win *pixelgl.Window, host string) *ConnectingScene {
	u := url.URL{Scheme: "ws", Host: host, Path: "/"}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Print("connected")
	conn := net.Websocket(c)
	scene := &ConnectingScene{
		next: make(chan Scene),
	}
	go func() {
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
			scene.next <- newMenu(win, conn)
			return
		}
	}()
	return scene
}
