package networking

import (
	"github.com/20zinnm/entity"
	"sync"
	"github.com/google/flatbuffers/go"
	"fmt"
	"github.com/20zinnm/spac/common/physics"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/server/health"
	"github.com/20zinnm/spac/server/shooting"
	"github.com/20zinnm/spac/server/movement"
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/20zinnm/spac/server/perceiving"
	"time"
	"math"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/builders"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/common/net/upstream"
	"sync/atomic"
	"github.com/20zinnm/spac/common/ship"
)

type networkingEntity struct {
	net.Connection
	filter      cp.ShapeFilter
	moveInputs  *moveInputs
	shootInputs *shootInputs
	known       map[entity.ID]struct{}
}

func sendSettings(conn net.Connection, radius float64) {
	b := builders.Get()
	defer builders.Put(b)
	downstream.ServerSettingsStart(b)
	downstream.ServerSettingsAddWorldRadius(b, radius)
	conn.Write(net.MessageDown(b, downstream.PacketServerSettings, downstream.ServerSettingsEnd(b)))
}

func sendSpawn(conn net.Connection, id entity.ID) {
	b := builders.Get()
	defer builders.Put(b)
	downstream.SpawnStart(b)
	downstream.SpawnAddId(b, id)
	conn.Write(net.MessageDown(b, downstream.PacketSpawn, downstream.SpawnEnd(b)))
}

func sendDeath(conn net.Connection) {
	b := builders.Get()
	defer builders.Put(b)
	downstream.DeathStart(b)
	conn.Write(net.MessageDown(b, downstream.PacketDeath, downstream.DeathEnd(b)))
}

type moveInputs struct {
	moveInputs *queue.RingBuffer
	lastMove   movement.Controls
}

func (m *moveInputs) Controls() movement.Controls {
	ctrls := m.lastMove
	if m.moveInputs.Len() > 0 {
		if i, err := m.moveInputs.Get(); err == nil {
			m.lastMove = i.(movement.Controls)
		}
	}
	return ctrls
}

type shootInputs struct {
	shootInputs *queue.RingBuffer
	lastShoot   shooting.Controls
}

func (e *shootInputs) Controls() shooting.Controls {
	ctrls := e.lastShoot
	if e.shootInputs.Len() > 0 {
		if i, err := e.shootInputs.Get(); err == nil {
			e.lastShoot = i.(shooting.Controls)
		}
	}
	return ctrls
}

type System struct {
	manager *entity.Manager
	world   *physics.World
	radius  float64
	// stateMu guards connections, entities, and lookup
	stateMu  sync.RWMutex
	entities map[entity.ID]*networkingEntity
	lookup   map[net.Connection]entity.ID
}

func New(manager *entity.Manager, world *physics.World, radius float64) *System {
	return &System{
		manager:  manager,
		world:    world,
		radius:   radius,
		lookup:   make(map[net.Connection]entity.ID),
		entities: make(map[entity.ID]*networkingEntity),
	}
}

func (s *System) Update(_ float64) {}

func (s *System) Remove(entity entity.ID) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if c, ok := s.entities[entity]; ok {
		go sendDeath(c)
		delete(s.entities, entity)
		delete(s.lookup, c)
	}
}

func (s *System) Add(conn net.Connection) {
	go func(conn net.Connection) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("encountered error decoding client message", err)
			}
			err := conn.Close()
			if err != nil {
				fmt.Print("closing conn:", err)
			}
			s.stateMu.Lock()
			if e, ok := s.lookup[conn]; ok {
				go s.manager.Remove(e)
				delete(s.entities, e)
			}
			delete(s.lookup, conn)
			s.stateMu.Unlock()
			fmt.Println("client disconnected")
		}()
		go sendSettings(conn, s.radius)
		for {
			data, err := conn.Read()
			if err != nil { // client disconnected
				return
			}
			message := upstream.GetRootAsMessage(data, 0)
			packetTable := new(flatbuffers.Table)
			if message.Packet(packetTable) {
				switch message.PacketType() {
				case upstream.PacketNONE:
				case upstream.PacketSpawn:
					play := new(upstream.Spawn)
					play.Init(packetTable.Bytes, packetTable.Pos)
					name := string(play.Name())
					if name == "" {
						name = "An Unnamed shipEntity"
					}
					newShip := shipEntity{
						ID:       s.manager.NewEntity(),
						Name:     name,
						Conn:     conn,
						Movement: new(movement.Controls),
						Shooting: &shooting.Component{
							Cooldown: 1 * time.Second,
						},
						Health: new(health.Component),
					}
					*newShip.Health = 100
					nete := &networkingEntity{
						Connection:  conn,
						known:       make(map[entity.ID]struct{}),
						moveInputs:  &moveInputs{moveInputs: queue.NewRingBuffer(4)},
						shootInputs: &shootInputs{shootInputs: queue.NewRingBuffer(4)},
						filter:      cp.NewShapeFilter(uint(newShip.ID), 0, cp.ALL_CATEGORIES),
					}
					s.world.Lock()
					newShip.Physics = ship.NewPhysics(s.world.Space, newShip.ID, cp.Vector{}) // todo spawn position should be random
					s.world.Unlock()
					for _, system := range s.manager.Systems() {
						switch sys := system.(type) {
						case *physics.System:
							sys.Add(newShip.ID, newShip.Physics)
						case *perceiving.System:
							sys.AddPerceiver(newShip.ID, &newShip)
							sys.AddPerceivable(newShip.ID, &newShip)
						case *movement.System:
							sys.Add(newShip.ID, nete.moveInputs, newShip.Physics, 100, 1)
						case *shooting.System:
							sys.Add(newShip.ID, nete.shootInputs, newShip.Physics, newShip.Shooting)
						case *health.System:
							sys.Add(newShip.ID, newShip.Health)
						}
					}
					s.stateMu.Lock()
					s.entities[newShip.ID] = nete
					s.lookup[conn] = newShip.ID
					s.stateMu.Unlock()
					go sendSpawn(conn, newShip.ID)
				case upstream.PacketControls:
					s.stateMu.RLock()
					if id, ok := s.lookup[conn]; ok {
						if e, ok := s.entities[id]; ok {
							controls := new(upstream.Controls)
							controls.Init(packetTable.Bytes, packetTable.Pos)
							e.moveInputs.moveInputs.Put(movement.Controls{controls.Left() > 0, controls.Right() > 0, controls.Thrusting() > 0})
						}
					}
					s.stateMu.RUnlock()
				}
			}
		}
	}(conn)
}

type shipEntity struct {
	ID       entity.ID
	Name     string
	Physics  physics.Component
	Conn     net.Connection
	Shooting *shooting.Component
	Health   *health.Component
	Movement *movement.Controls
}

func (s *shipEntity) Perceive(perception []byte) {
	s.Conn.Write(perception)
}

func (s *shipEntity) Position() cp.Vector {
	return s.Physics.Position()
}

func (s *shipEntity) Snapshot(builder *flatbuffers.Builder, known bool) flatbuffers.UOffsetT {
	var name *flatbuffers.UOffsetT
	if !known {
		name = new(flatbuffers.UOffsetT)
		*name = builder.CreateString(s.Name)
	}
	downstream.ShipStart(builder)
	posn := s.Position()
	downstream.ShipAddPosition(builder, downstream.CreateVector(builder, float32(posn.X), float32(posn.Y)))
	downstream.ShipAddRotation(builder, float32(s.Physics.Angle()))
	downstream.ShipAddHealth(builder, int16(math.Max(float64(atomic.LoadInt32((*int32)(s.Health))), 0)))
	if s.Shooting.Armed() {
		downstream.ShipAddArmed(builder, 1)
	} else {
		downstream.ShipAddArmed(builder, 0)
	}
	if s.Movement.Thrusting {
		downstream.ShipAddThrusting(builder, 1)
	} else {
		downstream.ShipAddThrusting(builder, 0)
	}
	if name != nil {
		downstream.ShipAddName(builder, *name)
	}
	ship := downstream.ShipEnd(builder)
	downstream.EntityStart(builder)
	downstream.EntityAddId(builder, s.ID)
	downstream.EntityAddSnapshotType(builder, downstream.SnapshotShip)
	downstream.EntityAddSnapshot(builder, ship)
	return downstream.EntityEnd(builder)
}
