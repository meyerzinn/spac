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
			s.manager.Remove(e.ID)
			delete(s.entities, e.ID)
		}
		delete(s.lookup, conn)
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
	networkingEntity := networkingEntity{
		ID:       id,
		controls: make(chan ship.Controls, 16),
	}
	// movement and shooting
	// thrusting indicates whether the ship is trying to thrust (purely cosmetic)--changing this DOES NOT actually make the ship move
	var thrusting bool
	movementQueue := make(chan movement.Controls, 16)
	shootingQueue := make(chan shooting.Controls, 16)
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
				thrusting = c.Movement.Thrusting
			}
		}
	}()
	// physics component
	s.world.Lock()
	physicsC := ship.Physics(s.world.Space, id) // todo spawn position should be random
	s.world.Unlock()
	// health component
	healthC := ship.Health()
	// shooting component
	shootingC := ship.Shooting()
	for _, system := range s.manager.Systems() {
		switch sys := system.(type) {
		case *physics.System:
			sys.Add(id, physicsC)
		case *perceiving.System:
			sys.AddPerceiver(id, ship.Perceiver(conn, physicsC))
			sys.AddPerceivable(id, ship.Perceivable(id, name, physicsC, healthC, shootingC, &thrusting))
		case *movement.System:
			sys.Add(id, movementQueue, physicsC, ship.LinearForce, ship.AngularVelocity)
		case *shooting.System:
			sys.Add(id, shootingC, shootingQueue, physicsC)
		case *health.System:
			sys.Add(id, healthC)
		}
	}
	// add entity to the networking index
	s.stateMu.Lock()
	s.entities[id] = conn
	s.lookup[conn] = networkingEntity
	s.stateMu.Unlock()
	// send spawn message
	b := builders.Get()
	defer builders.Put(b)
	downstream.SpawnStart(b)
	downstream.SpawnAddId(b, id)
	conn.Write(net.MessageDown(b, downstream.PacketSpawn, downstream.SpawnEnd(b)))
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

//
//func (s *System) handleSpawn(conn net.Connection, name string) {
//	var shipMu sync.RWMutex
//	name := string(play.Name())
//	if name == "" {
//		name = "An Unnamed shipEntity"
//	}
//	id := s.manager.NewEntity()
//	var movementC movement.Controls
//	movingQueue := make(chan movement.Controls, 16)
//	shootingC := &shooting.Component{
//		Cooldown:    1 * time.Second,
//		BulletForce: 300,
//	}
//	var shootingControls shooting.Controls
//	shootingQueue := make(chan shooting.Controls, 16)
//	healthC := ship.Health()
//	s.world.Lock()
//	physicsC := ship.Physics(s.world.Space, id) // todo spawn position should be random
//	s.world.Unlock()
//	for _, system := range s.manager.Systems() {
//		switch sys := system.(type) {
//		case *physics.System:
//			sys.Add(id, physicsC)
//		case *perceiving.System:
//			sys.AddPerceiver(id, &shipPerceiver{
//				conn:    conn,
//				physics: physicsC,
//			})
//			sys.AddPerceivable(id, ship.Perceivable(id, name, physicsC, healthC, shootingC, movementC))
//		case *movement.System:
//			sys.Add(id, movement.ControllerFunc(func() movement.Controls {
//
//			}), physicsC, movement.ShipVelocity, movement.ShipRotation)
//		case *shooting.System:
//			sys.Add(id, shooting.ControllerFunc(func() shooting.Controls {
//
//			}), physicsC, shootingC)
//		case *health.System:
//			sys.Add(id, healthC)
//		}
//	}
//	s.stateMu.Lock()
//	s.entities[id] = conn
//	s.lookup[conn] = id
//	s.movement[id] = movingQueue
//	s.shooting[id] = shootingQueue
//	s.stateMu.Unlock()
//	sendSpawn(conn, id)
//}
//
//func (s *System) handler(conn net.Connection) {
//	done := make(chan struct{})
//	defer func() {
//		close(done)
//		if err := recover(); err != nil {
//			fmt.Println("encountered error decoding client message", err)
//		}
//		err := conn.Close()
//		if err != nil {
//			fmt.Print("closing conn:", err)
//		}
//		s.stateMu.Lock()
//		if e, ok := s.lookup[conn]; ok {
//			s.manager.Remove(e)
//			delete(s.entities, e)
//		}
//		delete(s.lookup, conn)
//		s.stateMu.Unlock()
//		fmt.Println("client disconnected")
//	}()
//	sendSettings(conn, s.radius)
//	for {
//		data, err := conn.Read()
//		if err != nil { // client disconnected
//			return
//		}
//		message := upstream.GetRootAsMessage(data, 0)
//		packetTable := new(flatbuffers.Table)
//		if message.Packet(packetTable) {
//			switch message.PacketType() {
//			case upstream.PacketNONE:
//			case upstream.PacketPing:
//				ping := new(upstream.Ping)
//				ping.Init(packetTable.Bytes, packetTable.Pos)
//				b := builders.Get()
//				downstream.PongStart(b)
//				downstream.PongAddTimestamp(b, time.Now().UnixNano())
//				conn.Write(net.MessageDown(b, downstream.PacketPong, downstream.PongEnd(b)))
//				builders.Put(b)
//			case upstream.PacketSpawn:
//				play := new(upstream.Spawn)
//				play.Init(packetTable.Bytes, packetTable.Pos)
//				var shipMu sync.RWMutex
//				name := string(play.Name())
//				if name == "" {
//					name = "An Unnamed shipEntity"
//				}
//				id := s.manager.NewEntity()
//				var movementC movement.Controls
//				movingQueue := make(chan movement.Controls, 16)
//				shootingC := &shooting.Component{
//					Cooldown:    1 * time.Second,
//					BulletForce: 300,
//				}
//				var shootingControls shooting.Controls
//				shootingQueue := make(chan shooting.Controls, 16)
//				healthC := ship.Health()
//				//go func() {
//				//	queue := make(chan movement.Controls)
//				//	for {
//				//		select {
//				//		case <-done:
//				//			return
//				//		case c := <-queue:
//				//
//				//		}
//				//	}
//				//}()
//				//nete := &networkingEntity{
//				//	Connection:  conn,
//				//	known:       make(map[entity.ID]struct{}),
//				//	moveInputs:  &moveInputs{moveInputs: queue.NewRingBuffer(4), lastMove: movement},
//				//	shootInputs: &shootInputs{shootInputs: queue.NewRingBuffer(4)},
//				//	filter:      cp.NewShapeFilter(uint(id), 0, cp.ALL_CATEGORIES),
//				//}
//
//				//filter := cp.NewShapeFilter(uint(id), 0, cp.ALL_CATEGORIES)
//				s.world.Lock()
//				physicsC := ship.Physics(s.world.Space, id) // todo spawn position should be random
//				s.world.Unlock()
//				for _, system := range s.manager.Systems() {
//					switch sys := system.(type) {
//					case *physics.System:
//						sys.Add(id, physicsC)
//					case *perceiving.System:
//						sys.AddPerceiver(id, &shipPerceiver{
//							conn:    conn,
//							physics: physicsC,
//						})
//						sys.AddPerceivable(id, ship.Perceivable(id, name, physicsC, healthC, shootingC, movementC))
//					case *movement.System:
//						sys.Add(id, movement.ControllerFunc(func() movement.Controls {
//
//						}), physicsC, movement.ShipVelocity, movement.ShipRotation)
//					case *shooting.System:
//						sys.Add(id, shooting.ControllerFunc(func() shooting.Controls {
//
//						}), physicsC, shootingC)
//					case *health.System:
//						sys.Add(id, healthC)
//					}
//				}
//				s.stateMu.Lock()
//				s.entities[id] = conn
//				s.lookup[conn] = id
//				s.movement[id] = movingQueue
//				s.shooting[id] = shootingQueue
//				s.stateMu.Unlock()
//				sendSpawn(conn, id)
//			case upstream.PacketControls:
//				s.stateMu.RLock()
//				if id, ok := s.lookup[conn]; ok {
//					if e, ok := s.entities[id]; ok {
//						controls := new(upstream.Controls)
//						controls.Init(packetTable.Bytes, packetTable.Pos)
//						e.moveInputs.moveInputs.Put(movement.Controls{controls.Left() > 0, controls.Right() > 0, controls.Thrusting() > 0})
//					}
//				}
//				s.stateMu.RUnlock()
//			}
//		}
//	}
//}
//
//type shipPerceiver struct {
//	conn    net.Connection
//	physics world.Component
//}
//
//func (p *shipPerceiver) Position() cp.Vector {
//	return p.physics.Position()
//}
//
//func (p *shipPerceiver) Perceive(perception []byte) {
//	p.conn.Write(perception)
//}

//type shipEntity struct {
//	ID       entity.ID
//	Name     string
//	Physics  world.Component
//	Conn     net.Connection
//	Shooting *shooting.Component
//	Health   *health.Component
//	Movement *movement.Controls
//}
//
//func (s *shipEntity) Perceive(perception []byte) {
//	s.Conn.Write(perception)
//}
//
//func (s *shipEntity) Position() cp.Vector {
//	return s.Physics.Position()
//}
//
//func (s *shipEntity) Snapshot(builder *flatbuffers.Builder, known bool) flatbuffers.UOffsetT {
//	var name *flatbuffers.UOffsetT
//	if !known {
//		name = new(flatbuffers.UOffsetT)
//		*name = builder.CreateString(s.Name)
//	}
//	downstream.ShipStart(builder)
//	downstream.ShipAddPosition(builder, downstream.CreateVector(builder, float32(s.Position().X), float32(s.Position().Y)))
//	downstream.ShipAddVelocity(builder, downstream.CreateVector(builder, float32(s.Physics.Velocity().X), float32(s.Physics.Velocity().Y)))
//	downstream.ShipAddAngle(builder, float32(s.Physics.Angle()))
//	downstream.ShipAddAngularVelocity(builder, float32(s.Physics.AngularVelocity()))
//	downstream.ShipAddHealth(builder, int16(math.Max(float64(atomic.LoadInt32((*int32)(s.Health))), 0)))
//	if s.Shooting.Armed() {
//		downstream.ShipAddArmed(builder, 1)
//	} else {
//		downstream.ShipAddArmed(builder, 0)
//	}
//	if s.Movement.Thrusting {
//		downstream.ShipAddThrusting(builder, 1)
//	} else {
//		downstream.ShipAddThrusting(builder, 0)
//	}
//	if name != nil {
//		downstream.ShipAddName(builder, *name)
//	}
//	ship := downstream.ShipEnd(builder)
//	downstream.EntityStart(builder)
//	downstream.EntityAddId(builder, s.ID)
//	downstream.EntityAddSnapshotType(builder, downstream.SnapshotShip)
//	downstream.EntityAddSnapshot(builder, ship)
//	return downstream.EntityEnd(builder)
//}
