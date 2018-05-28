package ship

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/server/damaging"
	"github.com/20zinnm/spac/server/movement"
	"github.com/20zinnm/spac/server/physics/collision"
	"github.com/20zinnm/spac/server/shooting"
	"github.com/google/flatbuffers/go"
	"github.com/jakecoffman/cp"
	"math"
)

const (
	LinearForce     float64 = 400
	AngularVelocity float64 = 0.8
)

var shipVertices = []cp.Vector{{0, 51}, {-24, -21}, {0, -9}, {24, -21}}

type Controls struct {
	Movement movement.Controls
	Shooting shooting.Controls
}

type Entity struct {
	ID       entity.ID
	Conn     net.Connection
	Name     string
	Controls Controls
	Physics  *cp.Body
	Health   *damaging.Component
	Shooting *shooting.Component
	//sync.RWMutex // No need for mutex because there is no condition where an entity is updated and serialized simultaneously.
}

func New(space *cp.Space, id entity.ID, name string, conn net.Connection) *Entity {
	body := space.AddBody(cp.NewBody(1, cp.MomentForPoly(1, 4, shipVertices, cp.Vector{}, 0)))
	body.UserData = id
	shipShape := space.AddShape(cp.NewPolyShape(body, 4, shipVertices, cp.NewTransformIdentity(), 0))
	shipShape.SetFilter(cp.NewShapeFilter(uint(id), uint(collision.Damageable|collision.Perceivable), uint(collision.Damageable|collision.Perceiving)))
	return &Entity{
		ID:      id,
		Name:    name,
		Physics: body,
		Conn:    conn,
		Health: &damaging.Component{
			Value: 100,
			Max:   100,
		},
		Shooting: &shooting.Component{
			Cooldown:       20,
			BulletForce:    1000,
			BulletLifetime: 100,
		},
	}
}

func (s *Entity) Position() cp.Vector {
	return s.Physics.Position()
}

func (s *Entity) Perceive(perception []byte) {
	s.Conn.Write(perception)
}

func (s *Entity) Snapshot(b *flatbuffers.Builder, known bool) flatbuffers.UOffsetT {
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
	downstream.ShipAddHealth(b, int16(math.Max(s.Health.Value, 0)))
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
