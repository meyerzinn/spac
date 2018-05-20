package game

import (
	"context"
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/client/perceiving"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/upstream"
	"github.com/20zinnm/spac/common/net/builders"
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/spac/client/physics"
	"github.com/20zinnm/spac/client/rendering"
	"github.com/faiface/pixel/pixelgl"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/google/flatbuffers/go"
	"log"
)

type PlayingScene struct {
	ctx        context.Context
	manager    *entity.Manager
	perceiving *perceiving.System
	next       chan struct{}
}

func (s *PlayingScene) Update(dt float64) {
	select {
	case <-s.next:
		CurrentScene = NewMenuScene(s.ctx)
	}
}

func (s *PlayingScene) writer(queue chan rendering.Inputs) {
	conn := s.ctx.Value(CtxConnectionKey).(net.Connection)
	last := rendering.Inputs{}
	for {
		select {
		case <-s.next:
			return
		case <-s.ctx.Done():
			return
		case i := <-queue:
			if i != last {
				sendControls(conn, i)
				last = i
			}
		}
	}
}

func (s *PlayingScene) reader() {
	conn := s.ctx.Value(CtxConnectionKey).(net.Connection)
	for {
		select {
		case <-s.next:
			return
		case <-s.ctx.Done():
			return
		default:
			data, err := conn.Read()
			if err != nil {
				close(s.next)
				return
			}
			message := downstream.GetRootAsMessage(data, 0)
			packetTable := new(flatbuffers.Table)
			if !message.Packet(packetTable) {
				log.Println("error decoding message; skipping")
			}
			switch message.PacketType() {
			case downstream.PacketPerception:
				perception := new(downstream.Perception)
				perception.Init(packetTable.Bytes, packetTable.Pos)

			}
		}
	}
}

func NewPlayingScene(ctx context.Context) *PlayingScene {
	manager := new(entity.Manager)
	next := make(chan struct{})
	scene := &PlayingScene{
		ctx:     ctx,
		manager: manager,
		next:    next,
	}

	w := &world.World{Space: world.NewSpace()}
	manager.AddSystem(perceiving.New(manager, w))
	manager.AddSystem(physics.New(w))
	inputsQueue := make(chan rendering.Inputs, 16)
	manager.AddSystem(rendering.New(ctx.Value(CtxWindowKey).(*pixelgl.Window), w, rendering.InputHandlerFunc(func(i rendering.Inputs) {
		inputsQueue <- i
	})))
	go scene.writer(inputsQueue)

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
