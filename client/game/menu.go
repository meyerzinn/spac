package game

import (
	"context"
	"golang.org/x/image/colornames"
	"github.com/faiface/pixel/pixelgl"
	"github.com/20zinnm/spac/common/net"
	"fmt"
)

type MenuScene struct {
	ctx  context.Context
	name string
}

func NewMenuScene(ctx context.Context) Scene {
	return &MenuScene{
		ctx: ctx,
	}
}

func (m *MenuScene) Update(_ float64) {
	win := m.ctx.Value(CtxWindowKey).(*pixelgl.Window)
	win.Clear(colornames.Blue)
	m.name += win.Typed()
	if len(win.Typed()) > 0 {
	}
	if win.JustPressed(pixelgl.KeyDelete) {
		m.name = ""
	}
	if win.JustPressed(pixelgl.KeyEnter) {
		sendSpawn(m.ctx.Value(CtxConnectionKey).(net.Connection), m.name)
		CurrentScene = NewSpawningScene(m.ctx)
		fmt.Println("next scene (old:menu)")
	}
}
