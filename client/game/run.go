package game

import (
	"github.com/google/flatbuffers/go"
	"log"
	"github.com/20zinnm/spac/common/net/downstream"
	"time"
	"golang.org/x/image/colornames"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel"
	"net/url"
	"github.com/gorilla/websocket"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/upstream"
	"github.com/20zinnm/spac/common/net/builders"
	"github.com/20zinnm/spac/client/menu"
	"github.com/20zinnm/spac/client/playing"
)

func Run(host string) {
	// open window
	cfg := pixelgl.WindowConfig{
		Title:  "spac",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	// green = loading
	// todo: make actual loading screen
	win.Clear(colornames.Green)
	// connect to server
	u := url.URL{Scheme: "ws", Host: host, Path: "/"}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	conn := net.Websocket(c)
	in := make(chan *downstream.Message, 100)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			data, err := conn.Read()
			if err != nil {
				return
			}
			in <- downstream.GetRootAsMessage(data, 0)
		}
	}()
	msg := <-in
	if msg.PacketType() != downstream.PacketServerSettings {
		log.Fatal("first packet received from server is not settings")
	}
	packetTable := new(flatbuffers.Table)
	if !msg.Packet(packetTable) {
		log.Fatal("decode packet")
	}
	settings := new(downstream.ServerSettings)
	settings.Init(packetTable.Bytes, packetTable.Pos)
	radius := settings.WorldRadius()
	log.Print("world radius: ", radius)

	var scene Scene = menu.New(win, menu.HandlerFunc(func(name string) {
		log.Print("spawning...")
		sendSpawn(conn, name)
	}))
	last := time.Now()
	for !win.Closed() {
		now := time.Now()
		delta := last.Sub(now).Seconds()
		last = now
		scene.Update(delta)
		win.Update()
		select {
		case <-done:
			log.Fatal("lost connection to server")
		case message := <-in:
			packetTable := new(flatbuffers.Table)
			if !message.Packet(packetTable) {
				log.Fatal("parse message")
			}
			switch message.PacketType() {
			case downstream.PacketSpawn:
				log.Print("spawned")
				spawn := new(downstream.Spawn)
				spawn.Init(packetTable.Bytes, packetTable.Pos)
				scene = playing.New(win, radius, spawn.Id())
			case downstream.PacketPerception:
				perception := new(downstream.Perception)
				perception.Init(packetTable.Bytes, packetTable.Pos)
				scene.(*playing.Scene).Perceive(perception)
			}
		default:
			break
		}
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
