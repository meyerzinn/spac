package networking

import (
	"fmt"
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/common/constants"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/builders"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/common/net/upstream"
	"github.com/20zinnm/spac/server/bounding"
	"github.com/20zinnm/spac/server/entities/ship"
	"github.com/20zinnm/spac/server/health"
	"github.com/20zinnm/spac/server/movement"
	"github.com/20zinnm/spac/server/perceiving"
	"github.com/20zinnm/spac/server/physics"
	"github.com/20zinnm/spac/server/shooting"
	"github.com/google/flatbuffers/go"
	"github.com/jakecoffman/cp"
	"io"
	"math"
	"sync/atomic"
)

type Handler interface {
	Handle(message *upstream.Message)
}

func sendSettings(conn net.Connection, radius float64) {
	b := builders.Get()
	defer builders.Put(b)
	downstream.ServerSettingsStart(b)
	downstream.ServerSettingsAddWorldRadius(b, radius)
	conn.Write(net.MessageDown(b, downstream.PacketServerSettings, downstream.ServerSettingsEnd(b)))
}

func sendDeath(conn net.Connection) {
	b := builders.Get()
	defer builders.Put(b)
	downstream.DeathStart(b)
	conn.Write(net.MessageDown(b, downstream.PacketDeath, downstream.DeathEnd(b)))
}

type networkingEntity struct {
	entity.ID
	controls chan ship.Controls
}

type System struct {
	manager *entity.Manager
	space   *cp.Space
	radius  float64
	// instrumentation
	in      int64
	rate    uint64
	ticks   int64
	handled int64
	// all functions in updating are called at the start of the next tick, in order, to ensure synchronous access to resources.
	updating    chan func()
	connections map[net.Connection]struct{}
	entities    map[entity.ID]net.Connection
	lookup      map[net.Connection]networkingEntity
}

func New(manager *entity.Manager, space *cp.Space, radius float64) *System {
	return &System{
		manager:  manager,
		space:    space,
		radius:   radius,
		updating: make(chan func(), 64),
		entities: make(map[entity.ID]net.Connection),
		lookup:   make(map[net.Connection]networkingEntity),
	}
}

func (s *System) Update(dt float64) {
	s.ticks++
	in := atomic.SwapInt64(&s.in, 0)
	point := float64(in) / dt
	rate := math.Float64frombits(atomic.LoadUint64(&s.rate))
	atomic.StoreUint64(&s.rate, math.Float64bits(rate+point/float64(s.ticks)))
	for {
		select {
		case fn := <-s.updating:
			fn()
		default:
			return
		}
	}
}

func (s *System) Remove(entity entity.ID) {
	if c, ok := s.entities[entity]; ok {
		go sendDeath(c)
		if networkingEntity, ok := s.lookup[c]; ok {
			close(networkingEntity.controls)
			delete(s.lookup, c)
		}
		delete(s.entities, entity)
	}
}

func (s *System) Add(conn net.Connection) {
	atomic.AddInt64(&s.handled, 1)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("networking error", err)
		}
		fmt.Println("client disconnected")
		s.updating <- func() {
			if e, ok := s.lookup[conn]; ok {
				go s.manager.Remove(e.ID)
			}
		}
		err := conn.Close()
		if err != nil {
			fmt.Print("closing conn:", err)
		} else {
			fmt.Println("connection closed")
		}
		atomic.AddInt64(&s.handled, -1)
	}()
	fmt.Println("client connecting... sending server settings")
	sendSettings(conn, s.radius)
	fmt.Println("client connected")
	for {
		data, err := conn.Read()
		if err != nil { // client disconnected
			panic(err)
		}
		atomic.AddInt64(&s.in, int64(len(data)))
		message := upstream.GetRootAsMessage(data, 0)
		packetTable := new(flatbuffers.Table)
		if !message.Packet(packetTable) {
			panic("unable to decode packet from client")
		}
		switch message.PacketType() {
		case upstream.PacketNONE:
			fmt.Println("received empty packet from client")
		case upstream.PacketSpawn:
			spawn := new(upstream.Spawn)
			spawn.Init(packetTable.Bytes, packetTable.Pos)
			s.updating <- func() { s.handleSpawn(conn, spawn) }
		case upstream.PacketControls:
			controls := new(upstream.Controls)
			controls.Init(packetTable.Bytes, packetTable.Pos)
			s.updating <- func() { s.handleControls(conn, controls) }
		default:
			panic("received unknown packet from client")
		}
	}
}

func (s *System) handleSpawn(conn net.Connection, su *upstream.Spawn) {
	// name
	name := string(su.Name())
	if name == "" {
		name = "An Unnamed shipEntity"
	}
	// id
	id := s.manager.NewEntity()

	// send spawn message here so it gets to the client before a perception does
	b := builders.Get()
	defer builders.Put(b)
	downstream.SpawnStart(b)
	downstream.SpawnAddId(b, id)
	conn.Write(net.MessageDown(b, downstream.PacketSpawn, downstream.SpawnEnd(b)))

	networkingEntity := networkingEntity{
		ID:       id,
		controls: make(chan ship.Controls, 16),
	}
	movementQueue := make(chan movement.Controls, 16)
	shootingQueue := make(chan shooting.Controls, 16)

	e := ship.New(s.space, id, name, conn)
	// handle controls (split the ship controls channel into the movement and shooting control queues)
	go func() {
		var last ship.Controls
		for c := range networkingEntity.controls {
			if c.Shooting != last.Shooting {
				shootingQueue <- c.Shooting
				last.Shooting = c.Shooting
			}
			if c.Movement != last.Movement {
				movementQueue <- c.Movement
				last.Movement = c.Movement
			}
			s.updating <- func(c ship.Controls) func() {
				return func() {
					e.Controls = c
				}
			}(c)
		}
	}()
	for _, system := range s.manager.Systems() {
		switch sys := system.(type) {
		case *physics.System:
			sys.Add(id, e.Physics)
		case *perceiving.System:
			sys.AddPerceiver(id, e)
			sys.AddPerceivable(id, e)
		case *movement.System:
			sys.Add(id, movementQueue, e.Physics, constants.ShipLinearForce, constants.ShipAngularForce)
		case *shooting.System:
			sys.Add(id, e.Shooting, shootingQueue, e.Physics)
		case *health.System:
			sys.Add(id, e.Health)
		case *bounding.System:
		}
	}
	// add e to the networking index
	s.entities[id] = conn
	s.lookup[conn] = networkingEntity
}

func (s *System) handleControls(conn net.Connection, controls *upstream.Controls) {
	if e, ok := s.lookup[conn]; ok {
		e.controls <- ship.Controls{
			Shooting: shooting.Controls{
				Shooting: controls.Shooting(),
			},
			Movement: movement.Controls{
				Left:      controls.Left(),
				Right:     controls.Right(),
				Thrusting: controls.Thrusting(),
			},
		}
	}
}

func (s *System) Debug(w io.Writer) {
	fmt.Fprintln(w, "networking")
	fmt.Fprintf(w, "connections=%d\n", atomic.LoadInt64(&s.handled))
	fmt.Fprintf(w, "bytes in per second=%f\n", math.Float64frombits(atomic.LoadUint64(&s.rate)))
}
