package game

import (
	"fmt"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/builders"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/common/net/upstream"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/flatbuffers/go"
	"log"
)

type SpawningScene struct {
	win  *pixelgl.Window
	conn net.Connection
	next chan Scene
}

func NewSpawning(win *pixelgl.Window, conn net.Connection, name string) *SpawningScene {
	scene := &SpawningScene{win: win,
		conn: conn,
		next: make(chan Scene),
	}
	go func() {
		go sendSpawn(conn, name)
		for {
			message, err := readMessage(conn)
			if err != nil {
				log.Fatalln(err)
			}
			packetTable := new(flatbuffers.Table)
			if !message.Packet(packetTable) {
				log.Fatalln("failed to decode packet")
			}
			if message.PacketType() != downstream.PacketSpawn {
				fmt.Println("received packet other than spawn", message.PacketType(), downstream.PacketSpawn)
			}
			spawn := new(downstream.Spawn)
			spawn.Init(packetTable.Bytes, packetTable.Pos)
			scene.next <- NewPlaying(win, conn, spawn.Id())
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
