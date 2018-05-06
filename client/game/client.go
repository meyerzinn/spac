package game

import (
	"github.com/20zinnm/spac/common/net"
	"net/url"
	"log"
	"github.com/gorilla/websocket"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/entity"
	"sync/atomic"
	"github.com/20zinnm/spac/common/physics/world"
	"github.com/20zinnm/spac/common/physics"
	"github.com/20zinnm/spac/client/inputs"
	"github.com/20zinnm/spac/common/net/builders"
	"github.com/20zinnm/spac/common/net/upstream"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel"
	"github.com/pkg/errors"
)

var (
	BufferedIn = 16
)

// Client is a network game client.
type Client struct {
	conn net.Connection
	in   chan []byte
	done chan struct{}

	window     *pixelgl.Window
	inputs     chan inputs.Controls
	lastInputs inputs.Controls

	worldRadius float64
	connected   int32
	playing     int32

	manager *entity.Manager
}

func NewClient(host string) (*Client) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	client := &Client{
		conn: net.Websocket(c),
		in:   make(chan []byte, BufferedIn),
		done: make(chan struct{}),
	}
	go client.reader()
	go client.handler()

	return client
}

func (c *Client) handler() {
	for {
		select {
		case <-c.done:
			return
		case data := <-c.in:
			message := downstream.GetRootAsMessage(data, 0)
			packetTable := new(flatbuffers.Table)
			if message.Packet(packetTable) {
				switch message.PacketType() {
				case downstream.PacketServerSettings:
					serverSettings := new(downstream.ServerSettings)
					serverSettings.Init(packetTable.Bytes, packetTable.Pos)
					c.worldRadius = serverSettings.WorldRadius()
					if !atomic.CompareAndSwapInt32(&c.connected, 0, 1) {
						panic("received server settings multiple times")
					}
				case downstream.PacketSpawn:
					if !atomic.CompareAndSwapInt32(&c.playing, 0, 1) {
						panic("received spawn packet while playing")
					}
					cfg := pixelgl.WindowConfig{
						Title:  "spac",
						Bounds: pixel.R(0, 0, 1024, 768),
						VSync:  true,
					}
					win, err := pixelgl.NewWindow(cfg)
					if err != nil {
						panic(errors.Wrap(err, "could not open window"))
					}
					c.window = win
					c.manager = new(entity.Manager)
					world := world.New(physics.NewSpace())
					c.manager.AddSystem(physics.New(c.manager, world, c.worldRadius))
					c.manager.AddSystem(inputs.New(c.window, inputs.HandlerFunc(c.handleInputs)))
				}
			}
		}
	}
}

func (c *Client) handleInputs(in inputs.Controls) {
	c.inputs <- in
}

func (c *Client) writer() {
	for {
		select {
		case <-c.done:
			return
		case in := <-c.inputs:
			b := builders.Get()
			upstream.ControlsStart(b)
			upstream.ControlsAddLeft(b, boolToByte(in.Left))
			upstream.ControlsAddRight(b, boolToByte(in.Right))
			upstream.ControlsAddThrusting(b, boolToByte(in.Thrust))
			upstream.ControlsAddShooting(b, boolToByte(in.Shoot))
			go c.conn.Write(net.MessageUp(b, upstream.PacketControls, upstream.ControlsEnd(b)))
		}
	}
}

func (c *Client) reader() {
	defer func() {
		close(c.done)
		c.conn.Close()
	}()
	for {
		data, err := c.conn.Read()
		if err != nil {
			return
		}
		c.in <- data
	}
}

func boolToByte(val bool) byte {
	if val {
		return 1
	} else {
		return 0
	}
}
