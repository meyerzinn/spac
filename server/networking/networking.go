package networking

import (
	"github.com/20zinnm/entity"
	"sync"
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/builders"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/common/net/upstream"
	"github.com/google/flatbuffers/go"
	"fmt"
	"time"
	"github.com/20zinnm/spac/server/entities/ship"
	"github.com/20zinnm/spac/server/movement"
	"github.com/20zinnm/spac/server/shooting"
	"github.com/20zinnm/spac/server/health"
	"github.com/20zinnm/spac/server/physics"
	"github.com/20zinnm/spac/server/perceiving"
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
	world   *world.World
	radius  float64
	// stateMu guards connections, entities, and lookup
	stateMu     sync.RWMutex
	connections map[net.Connection]struct{}
	entities    map[entity.ID]net.Connection
	lookup      map[net.Connection]networkingEntity
	//movement map[entity.ID]chan movement.Controls
	//shooting map[entity.ID]chan shooting.Controls
}

func New(manager *entity.Manager, world *world.World, radius float64) *System {
	return &System{
		manager:  manager,
		world:    world,
		radius:   radius,
		entities: make(map[entity.ID]net.Connection),
		lookup:   make(map[net.Connection]networkingEntity),
	}
}

func (s *System) Update(_ float64) {
}

func (s *System) Remove(entity entity.ID) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
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
	go s.handle(conn)
}

func (s *System) handle(conn net.Connection) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("networking error", err)
		}
		fmt.Println("client disconnected")
		s.stateMu.Lock()
		if e, ok := s.lookup[conn]; ok {
			go s.manager.Remove(e.ID)
		}
		s.stateMu.Unlock()
		err := conn.Close()
		if err != nil {
			fmt.Print("closing conn:", err)
		} else {
			fmt.Println("connection closed")
		}
	}()
	fmt.Println("client connecting... sending server settings")
	sendSettings(conn, s.radius)
	fmt.Println("client connected")
	for {
		data, err := conn.Read()
		if err != nil { // client disconnected
			panic(err)
		}
		message := upstream.GetRootAsMessage(data, 0)
		packetTable := new(flatbuffers.Table)
		if !message.Packet(packetTable) {
			panic("unable to decode packet from client")
		}
		switch message.PacketType() {
		case upstream.PacketNONE:
			fmt.Println("received empty packet from client")
		case upstream.PacketPing:
			ping := new(upstream.Ping)
			ping.Init(packetTable.Bytes, packetTable.Pos)
			s.handlePing(conn, ping)
		case upstream.PacketSpawn:
			spawn := new(upstream.Spawn)
			spawn.Init(packetTable.Bytes, packetTable.Pos)
			s.handleSpawn(conn, spawn)
		case upstream.PacketControls:
			controls := new(upstream.Controls)
			controls.Init(packetTable.Bytes, packetTable.Pos)
			s.handleControls(conn, controls)
		default:
			panic("received unknown packet from client")
		}
	}
}

func (s *System) handlePing(conn net.Connection, ping *upstream.Ping) {
	b := builders.Get()
	defer builders.Put(b)
	downstream.PongStart(b)
	downstream.PongAddTimestamp(b, time.Now().UnixNano())
	conn.Write(net.MessageDown(b, downstream.PacketPong, downstream.PongEnd(b)))
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

	entity := ship.New(s.world, id, name, conn)
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
			entity.Lock()
			entity.Controls = c
			entity.Unlock()
		}
	}()
	for _, system := range s.manager.Systems() {
		switch sys := system.(type) {
		case *physics.System:
			sys.Add(id, entity.Physics)
		case *perceiving.System:
			sys.AddPerceiver(id, entity)
			sys.AddPerceivable(id, entity)
		case *movement.System:
			sys.Add(id, movementQueue, entity.Physics, ship.LinearForce, ship.AngularVelocity)
		case *shooting.System:
			sys.Add(id, entity.Shooting, shootingQueue, entity.Physics)
		case *health.System:
			sys.Add(id, entity.Health)
		}
	}
	// add entity to the networking index
	s.stateMu.Lock()
	s.entities[id] = conn
	s.lookup[conn] = networkingEntity
	s.stateMu.Unlock()
}

func (s *System) handleControls(conn net.Connection, controls *upstream.Controls) {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	if e, ok := s.lookup[conn]; ok {
		e.controls <- ship.Controls{
			Shooting: shooting.Controls{
				Shooting: controls.Shooting() > 0,
			},
			Movement: movement.Controls{
				Left:      controls.Left() > 0,
				Right:     controls.Right() > 0,
				Thrusting: controls.Thrusting() > 0,
			},
		}
	}
}
