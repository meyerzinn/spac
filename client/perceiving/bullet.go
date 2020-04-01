package perceiving

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/client/physics"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	flatbuffers "github.com/google/flatbuffers/go"
	"image/color"
	"time"
)

type bulletPhysics struct {
	timestamp time.Time
	physics.TranslationalState
}

type Bullet struct {
	ID entity.ID
	physics.TranslationalState
	bufferLen int
	buffer    [InterpolationBuffer]bulletPhysics
}

func NewBullet(id entity.ID) *Bullet {
	return &Bullet{
		ID: id,
	}
}

func (b *Bullet) FixedUpdate() {
	interpolationTime := time.Now().Add(-InterpolationBackTime)
	if b.buffer[0].timestamp.After(interpolationTime) {
		// INTERPOLATE
		for i := 1; i <= b.bufferLen; i++ {
			if b.buffer[i].timestamp.Before(interpolationTime) || i == b.bufferLen-1 {
				newer := b.buffer[i-1]
				older := b.buffer[i]

				length := newer.timestamp.Sub(older.timestamp).Seconds()
				var t float64
				if length > 0.0001 {
					t = interpolationTime.Sub(older.timestamp).Seconds() / length
				}
				b.TranslationalState = b.Lerp(older.Lerp(newer.TranslationalState, t), InterpolationBackTime.Seconds())
				return
			}

		}
	} else {
		// EXTRAPOLATE (rough)
		dt := time.Now().Sub(b.buffer[0].timestamp).Seconds()
		newer := b.Step(dt)
		b.TranslationalState = b.TranslationalState.Lerp(b.buffer[0].Lerp(newer, InterpolationBackTime.Seconds()), InterpolationConstant)
	}
}

func (b *Bullet) Draw(_ *pixelgl.Canvas, imd *imdraw.IMDraw) {

	imd.Color = color.RGBA{
		R: 74,
		G: 136,
		B: 212,
		A: 255,
	}

	imd.Push(pixel.Vec(b.Position))
	imd.Circle(8, 0)
}

func (b *Bullet) Update(timestamp time.Time, table *flatbuffers.Table) {
	bullet := new(downstream.Bullet)
	bullet.Init(table.Bytes, table.Pos)
	copy(b.buffer[1:], b.buffer[:])
	b.buffer[0] = bulletPhysics{
		timestamp: timestamp,
		TranslationalState: physics.TranslationalState{
			Position: tocpv(bullet.Position(new(downstream.Vector))),
			Velocity: tocpv(bullet.Velocity(new(downstream.Vector))),
		},
	}
	if b.bufferLen == 0 {
		b.TranslationalState = b.buffer[0].TranslationalState
	}
	if b.bufferLen < InterpolationBuffer {
		b.bufferLen++
	}
	//if posn.DistanceSq(b.lastPos) < 100 {
	//	b.Physicb.SetPosition(b.lastPob.Lerp(posn, delta))
	//	b.Physicb.SetVelocityVector(b.lastVel.Lerp(vel, delta))
	//} else {
	//	b.Physicb.SetPosition(posn)
	//	b.Physicb.SetVelocityVector(vel)
	//}
	//b.lastPos = posn
	//b.lastVel = vel
}
