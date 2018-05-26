package game

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel"
	"os/signal"
	"os"
	"time"
	"fmt"
)

func Run(host string) {
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
	CurrentScene = newConnecting(win, host)
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
