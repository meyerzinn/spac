package shooting

import "time"

type Component struct {
	Cooldown time.Duration
	LastShot time.Time
	BulletVelocity float64
}

func (c Component) Armed() bool {
	return c.LastShot.Add(c.Cooldown).Nanosecond() < time.Now().Nanosecond()
}

type Controls struct {
	Shooting bool
}

type Controller interface {
	Controls() Controls
}
