package networking

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/client/rendering"
)

type System struct {
	conn     net.Connection
	manager  *entity.Manager
	renderer rendering.Renderer
}

func New(conn net.Connection, manager *entity.Manager, renderer rendering.Renderer) *System {
	system := &System{
		conn:     conn,
		manager:  manager,
		renderer: renderer,
	}
	go system.loop()
	return system
}

func (s *System) loop() {
	for {
		data, err := s.conn.Read()
		if err != nil {

			return
		}

	}
}

func (s *System) Update(delta float64) {
	panic("implement me")
}

func (s *System) Remove(entity entity.ID) {
	panic("implement me")
}
