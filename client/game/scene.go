package game

type Manager struct {

}

type Scene interface {
	Update(dt float64)
}

var CurrentScene Scene