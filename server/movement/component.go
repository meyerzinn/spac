package movement

type Controls struct {
	Left, Right, Thrusting bool
}

type Controller interface {
	Controls() Controls
}
