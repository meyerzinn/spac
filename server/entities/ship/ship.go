package ship

import (
	"github.com/20zinnm/spac/server/shooting"
	"github.com/20zinnm/spac/common/net"
	"github.com/jakecoffman/cp"
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/common/net/downstream"
	"math"
	"github.com/20zinnm/spac/server/movement"
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/spac/server/health"
	"github.com/20zinnm/entity"
	"sync"
	"github.com/20zinnm/spac/server/physics/collision"
)

const (
	LinearForce     float64 = 400
	AngularVelocity float64 = 1
)

var shipVertices = []cp.Vector{{0, 51}, {-24, -21}, {0, -9}, {24, -21}}

type Controls struct {
	Movement movement.Controls
	Shooting shooting.Controls
}

type Entity struct {
	ID       entity.ID
	Name     string
	Controls Controls
	Physics  world.Component
	Conn     net.Connection
	Health   *health.Component
	Shooting *shooting.Component
	sync.RWMutex
}

func New(w *world.World, id entity.ID, name string, conn net.Connection) *Entity {
	w.Lock()
	body := w.Space.AddBody(cp.NewBody(1, cp.MomentForPoly(1, 4, shipVertices, cp.Vector{}, 0)))
	body.UserData = id
	shipShape := w.Space.AddShape(cp.NewPolyShape(body, 4, shipVertices, cp.NewTransformIdentity(), 0))
	shipShape.SetFilter(cp.NewShapeFilter(uint(id), uint(collision.Health|collision.Perceiving), cp.ALL_CATEGORIES))
	physics := world.Component{Body: body}
	w.Unlock()
	var health health.Component
	health = 100
	return &Entity{
		ID:      id,
		Name:    name,
		Physics: physics,
		Conn:    conn,
		Health:  &health,
		Shooting: &shooting.Component{
			Cooldown:       20,
			BulletForce:    1000,
			BulletLifetime: 100,
		},
	}
}

func (s *Entity) Position() cp.Vector {
	s.RLock()
	defer s.RUnlock()
	return s.Physics.Position()
}

func (s *Entity) Perceive(perception []byte) {
	s.Conn.Write(perception)
}

func (s *Entity) Snapshot(b *flatbuffers.Builder, known bool) flatbuffers.UOffsetT {
	s.RLock()
	defer s.RUnlock()
	var n *flatbuffers.UOffsetT
	if !known {
		n = new(flatbuffers.UOffsetT)
		*n = b.CreateString(s.Name)
	}
	downstream.ShipStart(b)
	downstream.ShipAddPosition(b, downstream.CreateVector(b, float32(s.Physics.Position().X), float32(s.Physics.Position().Y)))
	downstream.ShipAddVelocity(b, downstream.CreateVector(b, float32(s.Physics.Velocity().X), float32(s.Physics.Velocity().Y)))
	downstream.ShipAddAngle(b, float32(s.Physics.Angle()))
	downstream.ShipAddAngularVelocity(b, float32(s.Physics.AngularVelocity()))
	downstream.ShipAddHealth(b, int16(math.Max(float64(*s.Health), 0)))
	if s.Shooting.Armed() {
		downstream.ShipAddArmed(b, 1)
	} else {
		downstream.ShipAddArmed(b, 0)
	}
	if s.Controls.Movement.Thrusting {
		downstream.ShipAddThrusting(b, 1)
	} else {
		downstream.ShipAddThrusting(b, 0)
	}
	if n != nil {
		downstream.ShipAddName(b, *n)
	}
	ship := downstream.ShipEnd(b)
	downstream.EntityStart(b)
	downstream.EntityAddId(b, s.ID)
	downstream.EntityAddSnapshotType(b, downstream.SnapshotShip)
	downstream.EntityAddSnapshot(b, ship)
	return downstream.EntityEnd(b)
}
