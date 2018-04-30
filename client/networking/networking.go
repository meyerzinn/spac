package networking

import (
	"github.com/20zinnm/entity"
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/gorilla/websocket"
	"os"
	"os/signal"
	"net/url"
	"log"
	"github.com/20zinnm/spac/server/networking/fbs"
	"github.com/20zinnm/spac/common/physics"
	"github.com/google/flatbuffers/go"
)

type ship struct {
	physics.Component
}

type Entity interface {
	Update(*fbs.Entity)
}

type System struct {
	inputs   *queue.RingBuffer
	entities map[entity.ID]Entity
}

func New(host string) *System {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: host, Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, data, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			message := fbs.GetRootAsMessage(data, 0)
			packetTable := new(flatbuffers.Table)
			if !message.Packet(packetTable) {

			}
		}
	}()
}

func (s *System) Update(delta float64) {

}

func (s *System) Remove(entity entity.ID) {
	return
}
