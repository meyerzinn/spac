package game

import (
	"github.com/20zinnm/spac/common/net"
	"context"
	"log"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/google/flatbuffers/go"
	"fmt"
)

var CtxTargetIDKey = "target"

type SpawningScene struct {
	ctx  context.Context
	conn net.Connection
	next chan Scene
}

func NewSpawningScene(ctx context.Context) *SpawningScene {
	scene := &SpawningScene{ctx: ctx, next: make(chan Scene)}
	go func() {
		conn := ctx.Value(CtxConnectionKey).(net.Connection)
		for {
			data, err := conn.Read()
			if err != nil {
				log.Fatalln("error reading from server", err)
			}
			message := downstream.GetRootAsMessage(data, 0)
			if message.PacketType() != downstream.PacketSpawn {
				log.Println("received packet other than spawn at spawning scene; discarding")
				continue
			}
			packetTable := new(flatbuffers.Table)
			if !message.Packet(packetTable) {
				log.Fatalln("failed to parse spawn message")
			}
			spawn := new(downstream.Spawn)
			spawn.Init(packetTable.Bytes, packetTable.Pos)
			scene.next <- NewPlayingScene(context.WithValue(scene.ctx, CtxTargetIDKey, spawn.Id()))
			return
		}
	}()
	return scene
}

func (s *SpawningScene) Update(_ float64) {
	select {
	case scene := <-s.next:
		fmt.Println("next scene")
		CurrentScene = scene
	default:
	}
}
