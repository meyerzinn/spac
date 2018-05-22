package game

import (
	"context"
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/client/perceiving"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/upstream"
	"github.com/20zinnm/spac/common/net/builders"
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/spac/client/rendering"
	"github.com/faiface/pixel/pixelgl"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/google/flatbuffers/go"
	"log"
	"os"
	"fmt"
	"time"
	"sync/atomic"
	"github.com/20zinnm/spac/client/physics"
)

type PlayingScene struct {
	ctx        context.Context
	manager    *entity.Manager
	perceiving *perceiving.System
	next       chan Scene
	done       chan struct{}
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
	default:
		lastPerception := atomic.LoadInt64(&s.lastPerception)
		//latency := atomic.LoadInt64(&s.latency)
		//delta := math.Max(time.Now().Sub(time.Unix(0, lastPerception+latency)).Seconds(), 0)
		//fmt.Println(lastPerception, latency, delta)
		//s.manager.Update(delta)
		delta := time.Now().Sub(time.Unix(0, lastPerception)).Seconds()
		s.manager.Update(delta)
	}
}

func (s *PlayingScene) writer(queue chan rendering.Inputs) {
	conn := s.ctx.Value(CtxConnectionKey).(net.Connection)
	last := rendering.Inputs{}
	pinger := time.NewTicker(time.Second)
	defer pinger.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-s.ctx.Done():
			return
		case i := <-queue:
			if i != last {
				sendControls(conn, i)
				fmt.Println("sending controls")
				last = i
			}
		case <-pinger.C:
			sendPing(conn, time.Now().UnixNano())
		}
	}
}

func (s *PlayingScene) reader() {
	conn := s.ctx.Value(CtxConnectionKey).(net.Connection)
	defer close(s.done)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			data, err := conn.Read()
			if err != nil {
				log.Println("disconnected")
				os.Exit(0)
				// todo just go back to connecting and try again instead of exiting
				return
			}
			message := downstream.GetRootAsMessage(data, 0)
			packetTable := new(flatbuffers.Table)
			if !message.Packet(packetTable) {
				log.Println("error decoding message; skipping")
			}
			switch message.PacketType() {
			case downstream.PacketPong:
				pong := new(downstream.Pong)
				pong.Init(packetTable.Bytes, packetTable.Pos)
				//atomic.SwapInt64(&s.latency, time.Now().UnixNano()-pong.Timestamp())
			case downstream.PacketPerception:
				perception := new(downstream.Perception)
				perception.Init(packetTable.Bytes, packetTable.Pos)
				atomic.SwapInt64(&s.lastPerception, time.Now().UnixNano())
				s.perceiving.Perceive(perception)
			case downstream.PacketDeath:
				s.next <- NewMenuScene(s.ctx)
				return
			}
		}
	}
}

func NewPlayingScene(ctx context.Context) *PlayingScene {
	manager := new(entity.Manager)
	scene := &PlayingScene{
		ctx:     ctx,
		manager: manager,
		next:    make(chan Scene),
		done:    make(chan struct{}),
	}
	w := &world.World{Space: world.NewSpace()}
	scene.perceiving = perceiving.New(manager, w, ctx.Value(CtxTargetIDKey).(entity.ID), /*&scene.latency*/)
	manager.AddSystem(scene.perceiving)
	manager.AddSystem(physics.New(w))
	inputsQueue := make(chan rendering.Inputs, 16)
	manager.AddSystem(rendering.New(ctx.Value(CtxWindowKey).(*pixelgl.Window), w, rendering.InputHandlerFunc(func(i rendering.Inputs) {
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
	upstream.ControlsAddLeft(b, boolToByte(inputs.Left))
	upstream.ControlsAddRight(b, boolToByte(inputs.Right))
	upstream.ControlsAddThrusting(b, boolToByte(inputs.Thrust))
	upstream.ControlsAddShooting(b, boolToByte(inputs.Shoot))
	conn.Write(net.MessageUp(b, upstream.PacketControls, upstream.ControlsEnd(b)))
}

func sendPing(conn net.Connection, time int64) {
	b := builders.Get()
	defer builders.Put(b)
	upstream.PingStart(b)
	upstream.PingAddTimestamp(b, time)
	conn.Write(net.MessageUp(b, upstream.PacketPing, upstream.PingEnd(b)))
}
