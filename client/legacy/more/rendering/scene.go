package rendering

type Scene interface {
	Update(dt float64)
	Render()
	Destroy() // called once to unload stuff
}

var CurrentScene Scene