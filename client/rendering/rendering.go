package rendering

import (
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/client/stars"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"image/color"
	"math"
)

type Target interface {
}

type System struct {
	win            *pixelgl.Window
	handler        InputHandler
	entities       map[entity.ID]Renderable
	camPos         pixel.Vec
	camera         Camera
	canvas         *pixelgl.Canvas
	imd            *imdraw.IMDraw
}

func New(win *pixelgl.Window, handler InputHandler) *System {
	return &System{
		win:            win,
		handler:        handler,
		entities:       make(map[entity.ID]Renderable),
		canvas:         pixelgl.NewCanvas(win.Bounds().Moved(win.Bounds().Center().Scaled(-1))),
		imd:            imdraw.New(nil),
	}
}

func (s *System) Add(entity entity.ID, renderable Renderable) {
	s.entities[entity] = renderable
}

func (s *System) SetCamera(camera Camera) {
	s.camera = camera
}

func (s *System) Update(delta float64) {
	lerp := 1 - math.Pow(0.1, delta)

	var targetPosn pixel.Vec
	var health int
	if s.camera != nil {
		targetPosn = s.camera.Position()
		health = s.camera.Health()
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
	s.camPos = pixel.Lerp(s.camPos, targetPosn, lerp)
	cam := pixel.IM.Moved(s.camPos.Scaled(-1))
	s.canvas.SetMatrix(cam)
	s.canvas.Clear(colornames.Black)
	s.imd.Clear()
	s.imd.Color = colornames.Darkgray
	stars.Draw(s.imd, s.camPos, s.canvas.Bounds(), 8)
	s.imd.Color = colornames.Gray
	stars.Draw(s.imd, s.camPos, s.canvas.Bounds(), 4)
	s.imd.Color = colornames.White
	stars.Draw(s.imd, s.camPos, s.canvas.Bounds(), 2)

	for _, e := range s.entities {
		e.Draw(s.canvas, s.imd)
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

var healthOverlayVertices = []pixel.Vec{{256, 15}, {256, 45}, {768, 45}, {768, 15}}
var healthOverlayFullLength = 512

func drawHealthOverlay(imd *imdraw.IMDraw, bounds pixel.Rect, health int) {
	scale := bounds.W() / 1024
	filled := float64(health) / 100.0
	if health > 50 {
		imd.Color = color.RGBA{R: 46, G: 204, B: 113, A: 204}
	} else if health > 30 {
		imd.Color = color.RGBA{R: 211, G: 84, A: 255}
	} else {
		imd.Color = color.RGBA{R: 192, G: 57, B: 43, A: 255}
	}
	imd.Push(healthOverlayVertices[0].Scaled(scale))
	imd.Push(healthOverlayVertices[1].Scaled(scale))
	off := pixel.Vec{X: float64(healthOverlayFullLength) * filled}
	imd.Push(healthOverlayVertices[1].Add(off).Scaled(scale))
	imd.Push(healthOverlayVertices[0].Add(off).Scaled(scale))
	imd.Polygon(0)
	imd.Color = color.RGBA{R: 189, G: 195, B: 199, A: 204}
	for _, v := range healthOverlayVertices {
		imd.Push(v.Scaled(scale))
	}
	imd.Polygon(3)
}
