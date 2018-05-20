package game

import (
	"github.com/20zinnm/spac/common/net"
	"context"
	"github.com/20zinnm/spac/common/net/downstream"
	"log"
	"github.com/google/flatbuffers/go"
	"net/url"
	"github.com/gorilla/websocket"
	"fmt"
)

var CtxConnectionKey = "connection"
var CtxWorldRadiusKey = "worldRadius"

type ConnectingScene struct {
	ctx  context.Context
	next chan Scene
}

func (s *ConnectingScene) Update(dt float64) {
	select {
	case scene := <-s.next:
		CurrentScene = scene
	}
}

func NewConnecting(ctx context.Context, host string) *ConnectingScene {
	u := url.URL{Scheme: "ws", Host: host, Path: "/"}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Print("connected")
	conn := net.Websocket(c)
	ready := make(chan struct{})
	scene := &ConnectingScene{ctx: context.WithValue(ctx, CtxConnectionKey, conn)}
	go func() {
		for {
			select {
			case <-scene.ctx.Done():
				return
			}
			data, err := conn.Read()
			if err != nil {
				fmt.Println("disconnected", err)
				ctx.Done()
				return
			}
			message := downstream.GetRootAsMessage(data, 0)
			if message.PacketType() != downstream.PacketServerSettings {
				log.Println("received packet other than settings at connecting scene; discarding")
				continue
			}
			packetTable := new(flatbuffers.Table)
			if !message.Packet(packetTable) {
				log.Fatalln("failed to decode settings packet")
			}
			settings := new(downstream.ServerSettings)
			settings.Init(packetTable.Bytes, packetTable.Pos)
			scene.ctx = context.WithValue(scene.ctx, CtxWorldRadiusKey, settings.WorldRadius())
			close(ready)
		}
	}()
	return scene
}
