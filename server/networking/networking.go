package networking

import (
	"github.com/20zinnm/entity"
	"sync"
	"github.com/google/flatbuffers/go"
	"fmt"
	"github.com/20zinnm/spac/server/networking/fbs"
	"github.com/20zinnm/spac/common/physics"
	"github.com/20zinnm/spac/common/physics/world"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/server/health"
	"github.com/20zinnm/spac/server/shooting"
	"github.com/20zinnm/spac/server/movement"
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/20zinnm/spac/server/perceiving"
	"time"
	"math"
)

var (
	shipVertices = []cp.Vector{{-24, -20}, {24, -20}, {0, 40},}
)

type networkingEntity struct {
	Connection
	filter      cp.ShapeFilter
	moveInputs  *moveInputs
	shootInputs *shootInputs
	known       map[entity.ID]struct{}
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
	world   *world.World

	// stateMu guards connections, entities, and lookup
	stateMu  sync.RWMutex
	entities map[entity.ID]*networkingEntity
	lookup   map[Connection]entity.ID
}

func New(manager *entity.Manager, world *world.World) *System {
	return &System{
		manager:  manager,
		world:    world,
		lookup:   make(map[Connection]entity.ID),
		entities: make(map[entity.ID]*networkingEntity),
	}
}

func (s *System) Update(_ float64) {}

func (s *System) Remove(entity entity.ID) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if c, ok := s.entities[entity]; ok {
		delete(s.entities, entity)
		delete(s.lookup, c)
	}
}

func (s *System) Add(conn Connection) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()

	go func(conn Connection) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("encountered error decoding client message", err)
				conn.Close()
			}
		}()
		for {
			data, err := conn.Read()
			if err != nil { // client disconnected
				s.stateMu.Lock()
				if e, ok := s.lookup[conn]; ok {
					s.manager.Remove(e)
				}
				delete(s.lookup, conn)
				s.stateMu.Unlock()
				return
			}
			message := fbs.GetRootAsMessage(data, 0)
			packetTable := new(flatbuffers.Table)
			if message.Packet(packetTable) {
				switch message.PacketType() {
				case fbs.PacketNONE:
				case fbs.PacketSpawn:
					spawn := new(fbs.Spawn)
					spawn.Init(packetTable.Bytes, packetTable.Pos)
					name := string(spawn.Name())
					if name == "" {
						name = "An Unnamed Ship"
					}
					ship := ship{
						Name:     name,
						Conn:     conn,
						Movement: new(movement.Controls),
						Shooting: &shooting.Component{
							Cooldown: 1 * time.Second,
						},
					}
					ship.ID = s.manager.NewEntity()
					nete := &networkingEntity{
						Connection:  conn,
						known:       make(map[entity.ID]struct{}),
						moveInputs:  &moveInputs{moveInputs: queue.NewRingBuffer(4)},
						shootInputs: &shootInputs{shootInputs: queue.NewRingBuffer(4)},
						filter:      cp.NewShapeFilter(uint(ship.ID), 0, cp.ALL_CATEGORIES),
					}
					for _, system := range s.manager.Systems() {
						switch sys := system.(type) {
						case *physics.System:
							s.world.Do(func(space *cp.Space) {
								ship.Physics = physics.Component{Body: space.AddBody(cp.NewBody(1, cp.MomentForPoly(1, 3, shipVertices, cp.Vector{}, 0)))}
								shipShape := space.AddShape(cp.NewPolyShape(ship.Physics.Body, 3, shipVertices, cp.NewTransformIdentity(), 0))
								shipShape.SetFilter(cp.NewShapeFilter(uint(ship.ID), uint(perceiving.CollisionType), cp.ALL_CATEGORIES))
							})
							sys.Add(ship.ID, ship.Physics)
						case *perceiving.System:
							sys.AddPerceiver(ship.ID, &ship)
							sys.AddPerceivable(ship.ID, &ship)
						case *movement.System:
							sys.Add(ship.ID, nete.moveInputs, ship.Physics, 200, 0.03)
						case *shooting.System:
							sys.Add(ship.ID, nete.shootInputs, ship.Physics, ship.Shooting)
						}
					}
					s.stateMu.Lock()
					s.entities[ship.ID] = nete
					s.lookup[conn] = ship.ID
					s.stateMu.Unlock()
				case fbs.PacketControls:
					s.stateMu.RLock()
					if id, ok := s.lookup[conn]; ok {
						if e, ok := s.entities[id]; ok {
							controls := new(fbs.Controls)
							controls.Init(packetTable.Bytes, packetTable.Pos)
							e.moveInputs.moveInputs.Put(movement.Controls{controls.Left() > 0, controls.Right() > 0, controls.Thrusting() > 0})
						}
					}
				}
			}
		}
	}(conn)
}

type ship struct {
	ID       entity.ID
	Name     string
	Physics  physics.Component
	Conn     Connection
	Shooting *shooting.Component
	Health   *health.Component
	Movement *movement.Controls
}

func (s *ship) Perceive(perception []byte) {
	s.Conn.Write(perception)
}

func (s *ship) Position() cp.Vector {
	return s.Physics.Position()
}

func (s *ship) Snapshot(builder *flatbuffers.Builder, known bool) flatbuffers.UOffsetT {
	var name *flatbuffers.UOffsetT
	if !known {
		name = new(flatbuffers.UOffsetT)
		*name = builder.CreateString(s.Name)
	}
	fbs.ShipStart(builder)
	fbs.ShipAddId(builder, s.ID)
	posn := s.Position()
	fbs.ShipAddPosition(builder, fbs.CreatePoint(builder, int32(posn.X), int32(posn.Y)))
	fbs.ShipAddRotation(builder, float32(s.Physics.Angle()))
	fbs.ShipAddHealth(builder, int16(math.Max(float64(*s.Health), 0)))
	if s.Shooting.Armed() {
		fbs.ShipAddArmed(builder, 1)
	} else {
		fbs.ShipAddArmed(builder, 0)
	}
	if s.Movement.Thrusting {
		fbs.ShipAddThrusting(builder, 1)
	} else {
		fbs.ShipAddThrusting(builder, 0)
	}
	if name != nil {
		fbs.ShipAddName(builder, *name)
	}
	ship := fbs.ShipEnd(builder)
	fbs.EntityStart(builder)
	fbs.EntityAddSnapshotType(builder, fbs.SnapshotShip)
	fbs.EntityAddSnapshot(builder, ship)
	return fbs.EntityEnd(builder)
}
