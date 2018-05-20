package game

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/upstream"
	"github.com/20zinnm/spac/common/net/builders"
	"os/signal"
	"os"
	"context"
	"time"
	"fmt"
)

var CtxWindowKey = "window"

func Run(host string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// open window
	cfg := pixelgl.WindowConfig{
		Title:     "spac",
		Bounds:    pixel.R(0, 0, 1024, 768),
		VSync:     true,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	ctx = context.WithValue(ctx, CtxWindowKey, win)
	CurrentScene = NewConnecting(ctx, host)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ticker := time.NewTicker(time.Second / 60) // max 60 fps for now
	defer ticker.Stop()
	last := time.Now()
	for t := range ticker.C {
		select {
		case <-interrupt:
			return
		default:
			if win.Closed() {
				fmt.Println("window closed; exiting")
				os.Exit(0)
			}
			CurrentScene.Update(t.Sub(last).Seconds())
			last = t
			win.Update()
		}
	}
}

func boolToByte(val bool) byte {
	if val {
		return 1
	}
	return 0
}

func sendSpawn(conn net.Connection, name string) {
	b := builders.Get()
	defer builders.Put(b)
	nameOff := b.CreateString(name)
	upstream.SpawnStart(b)
	upstream.SpawnAddName(b, nameOff)
	conn.Write(net.MessageUp(b, upstream.PacketSpawn, upstream.SpawnEnd(b)))
}
