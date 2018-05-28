package rendering

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/client/stars"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/jakecoffman/cp"
	"golang.org/x/image/colornames"
	"image/color"
	"math"
	"fmt"
)

type System struct {
	win      *pixelgl.Window
	space    *cp.Space
	handler  InputHandler
	entities map[entity.ID]Renderable
	camPos   pixel.Vec
	tracking Trackable
	canvas   *pixelgl.Canvas
	imd      *imdraw.IMDraw
}

func New(win *pixelgl.Window, space *cp.Space, handler InputHandler) *System {
	return &System{
		win:      win,
		space:    space,
		handler:  handler,
		entities: make(map[entity.ID]Renderable),
		canvas:   pixelgl.NewCanvas(win.Bounds().Moved(win.Bounds().Center().Scaled(-1))),
		imd:      imdraw.New(nil),
	}
}

func (s *System) Add(entity entity.ID, renderable Renderable) {
	s.entities[entity] = renderable
}

func (s *System) Track(trackable Trackable) {
	s.tracking = trackable
}

var lastHealth int

func (s *System) Update(delta float64) {
	var targetPosn pixel.Vec
	var health int
	if s.tracking != nil {
		targetPosn = s.tracking.Position()
		health = s.tracking.Health()
		if health != lastHealth {
			fmt.Println(health)
			lastHealth = health
		}
	}
	inputs := Inputs{
		Left:   s.win.Pressed(pixelgl.KeyA),
		Right:  s.win.Pressed(pixelgl.KeyD),
		Thrust: s.win.Pressed(pixelgl.KeyW),
		Shoot:  s.win.Pressed(pixelgl.MouseButton1) || s.win.Pressed(pixelgl.KeySpace),
	}
	s.handler.Handle(inputs)
	if s.win.Bounds() != s.canvas.Bounds() {
		s.canvas.SetBounds(s.win.Bounds().Moved(s.win.Bounds().Center().Scaled(-1)))
	}
	s.camPos = pixel.Lerp(s.camPos, targetPosn, 1)
	cam := pixel.IM.Moved(s.camPos.Scaled(-1))
	s.canvas.SetMatrix(cam)
	s.canvas.Clear(colornames.Black)
	s.imd.Clear()
	s.imd.Color = colornames.Darkgray
	stars.Draw(s.imd, s.camPos, s.canvas.Bounds(), 4)
	s.imd.Color = colornames.Gray
	stars.Draw(s.imd, s.camPos, s.canvas.Bounds(), 2)
	s.imd.Color = colornames.White
	stars.Draw(s.imd, s.camPos, s.canvas.Bounds(), 1)
	for _, entity := range s.entities {
		entity.Draw(s.canvas, s.imd)
	}
	s.imd.Draw(s.canvas)

	// draw ui
	s.win.SetMatrix(pixel.IM)
	s.canvas.SetMatrix(pixel.IM.Moved(s.win.Bounds().Center().Scaled(-1)))
	s.imd.Clear()
	if health > 0 {
		drawHealthOverlay(s.imd, s.canvas.Bounds(), health)
	}
	s.imd.Draw(s.canvas)
	s.win.SetMatrix(pixel.IM.Scaled(pixel.ZV,
		math.Max(
			s.win.Bounds().W()/s.canvas.Bounds().W(),
			s.win.Bounds().H()/s.canvas.Bounds().H(),
		),
	))
	s.canvas.Draw(s.win, pixel.IM.Moved(s.win.Bounds().Center()))
}

func (s *System) Remove(entity entity.ID) {
	delete(s.entities, entity)
}

var healthOverlayVertices = []pixel.Vec{{256, 15}, {768, 15}, {768, 55}, {256, 55}}

//var healthOverlayInnerVertices = []pixel.Vec{{}}

func drawHealthOverlay(imd *imdraw.IMDraw, bounds pixel.Rect, health int) {
	//fill := health / 100
	scale := bounds.W() / 1024
	imd.Color = color.RGBA{R: 46, G: 204, B: 113, A: 204}
	ref := healthOverlayVertices[0]
	imd.Push(ref.Scaled(scale))
	imd.Push(ref.Add(pixel.V(float64(health/100*512), 0)).Scaled(scale))
	imd.Push(ref.Add(pixel.V(float64(health/100*510), 40)).Scaled(scale))
	imd.Push(ref.Add(pixel.V(0, 40)).Scaled(scale))
	imd.Polygon(0)
	imd.Color = color.RGBA{R: 189, G: 195, B: 199, A: 204}
	for _, v := range healthOverlayVertices {
		imd.Push(v.Scaled(scale))
	}
	imd.Polygon(3)
}
