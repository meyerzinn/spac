package perceiving

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/flatbuffers/go"
	"github.com/jakecoffman/cp"
	"image/color"
)

type Bullet struct {
	ID      entity.ID
	Physics *cp.Body
}

func NewBullet(space *cp.Space, id entity.ID) *Bullet {
	body := space.AddBody(cp.NewBody(1, cp.MomentForCircle(1, 0, 8, cp.Vector{})))
	space.AddShape(cp.NewCircle(body, 8, cp.Vector{}))
	return &Bullet{
		ID:      id,
		Physics: body,
	}
}

func (b *Bullet) Draw(_ *pixelgl.Canvas, imd *imdraw.IMDraw) {
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

func (b *Bullet) Update(table *flatbuffers.Table) {
	bullet := new(downstream.Bullet)
	bullet.Init(table.Bytes, table.Pos)
	posn := bullet.Position(new(downstream.Vector))
	vel := bullet.Velocity(new(downstream.Vector))
	b.Physics.SetPosition(cp.Vector{X: float64(posn.X()), Y: float64(posn.Y())})
	b.Physics.SetVelocityVector(cp.Vector{X: float64(vel.X()), Y: float64(vel.Y())})
}
