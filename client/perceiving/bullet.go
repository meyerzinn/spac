package perceiving

import (
	"github.com/20zinnm/spac/common/world"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/entity"
	"image/color"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"sync"
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/common/net/downstream"
)

type Bullet struct {
	ID      entity.ID
	Physics world.Component
	sync.RWMutex
}

func NewBullet(space *cp.Space, id entity.ID) *Bullet {
	body := space.AddBody(cp.NewBody(1, cp.MomentForCircle(1, 0, 8, cp.Vector{})))
	space.AddShape(cp.NewCircle(body, 8, cp.Vector{}))
	return &Bullet{
		ID:      id,
		Physics: world.Component{Body: body},
	}
}

func (b *Bullet) Draw(imd *imdraw.IMDraw) {
	b.RLock()
	defer b.RUnlock()
	imd.Color = color.RGBA{
		R: 74,
		G: 136,
		B: 212,
		A: 255,
	}
	posn := b.Physics.Position()
	imd.Push(pixel.Vec(posn))
	imd.Circle(8, 0)
}

func (b *Bullet) Update(bytes []byte, pos flatbuffers.UOffsetT) {
	bullet := new(downstream.Bullet)
	bullet.Init(bytes, pos)
	posn := bullet.Position(new(downstream.Vector))
	vel := bullet.Velocity(new(downstream.Vector))
	b.Lock()
	defer b.Unlock()
	b.Physics.SetPosition(cp.Vector{X: float64(posn.X()), Y: float64(posn.Y())})
	b.Physics.SetVelocityVector(cp.Vector{X: float64(vel.X()), Y: float64(vel.Y())})
}
