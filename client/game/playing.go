package game

import (
	"fmt"
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/client/perceiving"
	"github.com/20zinnm/spac/client/physics"
	"github.com/20zinnm/spac/client/rendering"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/builders"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/common/net/upstream"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/flatbuffers/go"
	"os"
	"sync/atomic"
	"time"
)

type PlayingScene struct {
	win        *pixelgl.Window
	conn       net.Connection
	manager    *entity.Manager
	perceiving *perceiving.System
	next       chan Scene
	//latency should only be modified atomically; it represents the nanosecond latency for a roundtrip packet
	//latency int64
	// lastPerception represents the time the lastPerception perception was received; it should only be modified atomically
	lastPerception int64
}

func (s *PlayingScene) Update(dt float64) {
	select {
	case next := <-s.next:
		fmt.Println("next scene (old:playing)")
		CurrentScene = next
		s.manager.Destroy()
	default:
		s.manager.Update(dt)
	}
}

func (s *PlayingScene) writer(queue chan rendering.Inputs) {
	last := rendering.Inputs{}
	for {
		select {
		case i, ok := <-queue:
			if !ok {
				return
			}
			if i != last {
				sendControls(s.conn, i)
				last = i
			}
		}
	}
}

func (s *PlayingScene) reader() {
	for {
		data, err := s.conn.Read()
		if err != nil {
			fmt.Println("disconnected")
			os.Exit(0)
			// todo just go back to connecting and try again instead of exiting
			return
		}
		message := downstream.GetRootAsMessage(data, 0)
		packetTable := new(flatbuffers.Table)
		if !message.Packet(packetTable) {
			fmt.Println("error decoding message; skipping")
		}
		switch message.PacketType() {
		case downstream.PacketPerception:
			perception := new(downstream.Perception)
			perception.Init(packetTable.Bytes, packetTable.Pos)
			atomic.SwapInt64(&s.lastPerception, time.Now().UnixNano())
			go s.perceiving.Perceive(perception)
		case downstream.PacketDeath:
			s.next <- NewDeath(s.win, s.conn)
			return
		}
	}
}

func NewPlaying(win *pixelgl.Window, conn net.Connection, self entity.ID) *PlayingScene {
	manager := entity.NewManager()
	scene := &PlayingScene{
		win:     win,
		conn:    conn,
		manager: manager,
		next:    make(chan Scene),
	}
	phys := physics.New()
	manager.AddSystem(phys)
	scene.perceiving = perceiving.New(manager, phys, self)
	manager.AddSystem(scene.perceiving)
	inputsQueue := make(chan rendering.Inputs, 16)
	manager.AddSystem(physics.New())
	manager.AddSystem(rendering.New(win, rendering.InputHandlerFunc(func(i rendering.Inputs) {
		inputsQueue <- i
	})))
	go scene.writer(inputsQueue)
	go scene.reader()
	return scene
}

func sendControls(conn net.Connection, inputs rendering.Inputs) {
	b := builders.Get()
	defer builders.Put(b)
	upstream.ControlsStart(b)
	upstream.ControlsAddLeft(b, inputs.Left)
	upstream.ControlsAddRight(b, inputs.Right)
	upstream.ControlsAddThrusting(b, inputs.Thrust)
	upstream.ControlsAddShooting(b, inputs.Shoot)
	conn.Write(net.MessageUp(b, upstream.PacketControls, upstream.ControlsEnd(b)))
}