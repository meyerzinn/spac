package game

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/client/inputs"
	"github.com/20zinnm/spac/common/physics"
	"github.com/20zinnm/spac/client/rendering"
	"github.com/faiface/pixel/pixelgl"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/fbs"
	"github.com/20zinnm/spac/common/net/builders"
	"github.com/20zinnm/spac/common/physics/world"
)

type PlayingScene struct {
	manager *entity.Manager
	window  *pixelgl.Window
	conn    net.Connection

	inputs   *inputs.System
	controls inputs.Controls

	physics  *physics.System
	renderer *rendering.GameRenderer
}

func (s *PlayingScene) handleInputs(c inputs.Controls) {
	if s.controls == c {
		return
	}
	s.controls = c
	b := builders.Get()
	fbs.ControlsStart(b)
	fbs.ControlsAddLeft(b, boolToByte(c.Left))
	fbs.ControlsAddRight(b, boolToByte(c.Right))
	fbs.ControlsAddThrusting(b, boolToByte(c.Thrust))
	fbs.ControlsAddShooting(b, boolToByte(c.Shoot))
	go s.conn.Write(net.Message(b, fbs.ControlsEnd(b), fbs.PacketControls))
}

func (s *PlayingScene) Destroy() {
	s.manager.RemoveSystem(s.inputs)
	s.manager.RemoveSystem(s.physics)
	s.manager.RemoveSystem(s.renderer)
}

func NewPlayingScene(manager *entity.Manager, window *pixelgl.Window) Scene {
	var scene PlayingScene
	scene.manager = manager
	scene.window = window

	inputs := inputs.New(window, inputs.HandlerFunc(scene.handleInputs))
	manager.AddSystem(inputs)

	space := physics.NewSpace()
	world := world.New(space)
	physics := physics.New(manager, world)
	manager.AddSystem(physics)
}

func boolToByte(val bool) byte {
	if val {
		return 1
	} else {
		return 0
	}
}
