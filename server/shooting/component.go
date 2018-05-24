package shooting

type Component struct {
	Cooldown uint
	// ticks to armed
	tta            uint
	BulletForce    float64
	BulletLifetime uint
}

func (c Component) Armed() bool {
	return c.tta == 0
}

type Controls struct {
	Shooting bool
}

type Controller chan Controls
