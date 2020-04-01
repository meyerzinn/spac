package server

import (
	"fmt"
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/spac/server/bounding"
	"github.com/20zinnm/spac/server/despawning"
	"github.com/20zinnm/spac/server/health"
	"github.com/20zinnm/spac/server/movement"
	"github.com/20zinnm/spac/server/networking"
	"github.com/20zinnm/spac/server/perceiving"
	"github.com/20zinnm/spac/server/physics"
	"github.com/20zinnm/spac/server/shooting"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"time"
)

type server struct {
	bind     string
	radius   float64
	tick     time.Duration
	upgrader *websocket.Upgrader
	debug    bool
}

type Option func(*server)

func BindAddress(address string) Option {
	return func(s *server) {
		s.bind = address
	}
}

func WorldRadius(radius float64) Option {
	return func(s *server) {
		s.radius = radius
	}
}

func TickRate(tick time.Duration) Option {
	return func(s *server) {
		s.tick = tick
	}
}

func Upgrader(upgrader *websocket.Upgrader) Option {
	return func(s *server) {
		s.upgrader = upgrader
	}
}

func Debug() Option {
	return func(s *server) {
		s.debug = true
	}
}

func Start(options ...Option) {
	server := &server{
		bind:   ":8080",
		radius: 10000,
		tick:   time.Second / 60,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	for _, o := range options {
		o(server)
	}
	var manager = entity.NewManager()
	defer manager.Destroy()
	space := world.NewSpace()
	manager.AddSystem(bounding.New(server.radius))
	manager.AddSystem(movement.New())
	manager.AddSystem(shooting.New(manager, space))
	manager.AddSystem(physics.New(manager, space))
	manager.AddSystem(perceiving.New(space))
	manager.AddSystem(health.New(manager, space))
	manager.AddSystem(despawning.New(manager))

	netwk := networking.New(manager, space, server.radius)
	manager.AddSystem(netwk)
	go func() {
		ticker := time.NewTicker(server.tick)
		defer ticker.Stop()
		last := time.Now()
		skip := 0
		for range ticker.C {
			if len(ticker.C) > 0 {
				skip++
				continue // we're behind!
			}
			if skip > 0 {
				fmt.Printf("Server lagging! Skipping %d ticks.", skip)
				skip = 0
			}
			now := time.Now()
			delta := now.Sub(last).Seconds()
			last = now
			manager.Update(delta)
		}
		fmt.Println("game stopped")

	}()
	fmt.Println("game started")
	http.Handle("/ws", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("error upgrading connection", err)
			return
		}
		netwk.Add(net.Websocket(conn))
	}))
	if server.debug {
		http.Handle("/debug", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "debugging")
			fmt.Fprintln(w, "=========")
			for _, system := range manager.Systems() {
				if debuggable, ok := system.(Debuggable); ok {
					debuggable.Debug(w)
					fmt.Fprintf(w, "---\n")
				}
			}
		}))
	}
	log.Fatal(http.ListenAndServe(server.bind, nil))
}

type Debuggable interface {
	Debug(to io.Writer)
}
