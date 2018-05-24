package ship

import (
	"github.com/20zinnm/spac/server/perceiving"
	"github.com/20zinnm/spac/common/world"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/common/net"
)

type shipPerceiver struct {
	conn    net.Connection
	physics world.Component
}

func (p *shipPerceiver) Position() cp.Vector {
	return p.physics.Position()
}

func (p *shipPerceiver) Perceive(perception []byte) {
	p.conn.Write(perception)
}

func Perceiver(conn net.Connection, physics world.Component) perceiving.Perceiver {
	return &shipPerceiver{
		conn:    conn,
		physics: physics,
	}
}
