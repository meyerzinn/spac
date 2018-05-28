package game

type Scene interface {
	Update(dt float64)
}

var CurrentScene Scene
