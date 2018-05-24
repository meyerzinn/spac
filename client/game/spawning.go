package game

import (
	"github.com/20zinnm/spac/common/net"
	"log"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/google/flatbuffers/go"
	"fmt"
	"github.com/faiface/pixel/pixelgl"
	"github.com/20zinnm/spac/common/net/upstream"
	"github.com/20zinnm/spac/common/net/builders"
)

type SpawningScene struct {
	win  *pixelgl.Window
	conn net.Connection
	next chan Scene
}

func newSpawning(win *pixelgl.Window, conn net.Connection, name string) *SpawningScene {
	scene := &SpawningScene{win: win,
		conn: conn,
		next: make(chan Scene),
	}
	go func() {
		sendSpawn(conn, name)
		for {
			message, err := readMessage(conn)
			if err != nil {
				log.Fatalln(err)
			}
			packetTable := new(flatbuffers.Table)
			if !message.Packet(packetTable) {
				log.Fatalln("failed to decode packet")
			}
			spawn := new(downstream.Spawn)
			spawn.Init(packetTable.Bytes, packetTable.Pos)
			scene.next <- newPlaying(win, conn, spawn.Id())
			return
		}
	}()
	return scene
}

func (s *SpawningScene) Update(_ float64) {
	select {
	case scene := <-s.next:
		fmt.Println("next scene (old:spawning)")
		CurrentScene = scene
	default:
	}
}

func sendSpawn(conn net.Connection, name string) {
	b := builders.Get()
	defer builders.Put(b)
	nameOff := b.CreateString(name)
	upstream.SpawnStart(b)
	upstream.SpawnAddName(b, nameOff)
	conn.Write(net.MessageUp(b, upstream.PacketSpawn, upstream.SpawnEnd(b)))
}
