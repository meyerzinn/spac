package rendering

type InputHandler interface {
	Handle(Inputs)
}

type InputHandlerFunc func(Inputs)

func (fn InputHandlerFunc) Handle(i Inputs) {
	fn(i)
}

type Inputs struct {
	Left, Right, Thrust, Shoot bool
}
