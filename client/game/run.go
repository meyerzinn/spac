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
	log.Print("connected")
	conn := net.Websocket(c)
	in := make(chan *downstream.Message, 100)
	disconnected := make(chan struct{})
	go func() {
		defer close(disconnected)
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

	controlsQueue := make(chan playing.Controls, 64)
	var controls playing.Controls

	go func() {
		for c := range controlsQueue {
			sendControls(conn, c)
		}
	}()

	var scene Scene = menu.New(win, menu.HandlerFunc(func(name string) {
		log.Print("spawning...")
		sendSpawn(conn, name)
	}))
	last := time.Now()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for t := range ticker.C {
		if win.Closed() {
			log.Print("window closed; exiting")
			return
		}
		delta := last.Sub(t).Seconds()
		last = t
		scene.Update(delta)
		win.Update()
		select {
		case <-disconnected:
			log.Fatal("disconnected from server")
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
				scene = playing.New(win, radius, spawn.Id(), playing.HandlerFunc(func(c playing.Controls) {
					if c != controls {
						controls = c
						controlsQueue <- c
					}
				}))
			case downstream.PacketPerception:
				perception := new(downstream.Perception)
				perception.Init(packetTable.Bytes, packetTable.Pos)
				scene.(*playing.Scene).Perceive(perception)
			case downstream.PacketDeath:
			}
		default:
			break
		}
	}
}

func boolToByte(val bool) byte {
	if val {
		return 1
	}
	return 0
}

func sendControls(conn net.Connection, controls playing.Controls) {
	b := builders.Get()
	defer builders.Put(b)
	upstream.ControlsStart(b)
	upstream.ControlsAddLeft(b, boolToByte(controls.Left))
	upstream.ControlsAddRight(b, boolToByte(controls.Right))
	upstream.ControlsAddThrusting(b, boolToByte(controls.Thrust))
	upstream.ControlsAddShooting(b, boolToByte(controls.Shoot))
	conn.Write(net.MessageUp(b, upstream.PacketControls, upstream.ControlsEnd(b)))
}

func sendSpawn(conn net.Connection, name string) {
	b := builders.Get()
	defer builders.Put(b)
	nameOff := b.CreateString(name)
	upstream.SpawnStart(b)
	upstream.SpawnAddName(b, nameOff)
	conn.Write(net.MessageUp(b, upstream.PacketSpawn, upstream.SpawnEnd(b)))
}
