package networking

import (
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/downstream"
)

type Handler interface {
	Handle(message *downstream.Message)
}

type HandlerFunc func(message *downstream.Message)

func (h HandlerFunc) Handle(message *downstream.Message) {
	h(message)
}

type System struct {
	conn    net.Connection
	handler Handler
}

func New(conn net.Connection) *System {
	system := &System{
		conn: conn,
	}
	go system.reader()
	return system
}

func (s *System) reader() {
	for {
		data, err := s.conn.Read()
		if err != nil {
			return
		}
		go s.handler.Handle(downstream.GetRootAsMessage(data, 0))
	}
}