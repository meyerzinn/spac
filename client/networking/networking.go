package networking

import (
	"github.com/20zinnm/entity"
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/20zinnm/spac/common/net/fbs"
	"github.com/20zinnm/spac/common/physics"
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/client/rendering"
	"sync"
	"github.com/jakecoffman/cp"
	"net/url"
	"log"
	"github.com/gorilla/websocket"
	"github.com/faiface/pixel/pixelgl"
	"github.com/20zinnm/spac/common/physics/world"
)

type shipEntity struct {
	ID        entity.ID
	Name      string
	Physics   physics.Component
	Rendering rendering.Component
	Armed     bool
	Thrusting bool
	Health    int
}

func (s *shipEntity) Update(e *fbs.Entity) {
	snapshotTable := new(flatbuffers.Table)
	if !e.Snapshot(snapshotTable) {
		return
	}
	var snapshot = new(fbs.Ship)
	snapshot.Init(snapshotTable.Bytes, snapshotTable.Pos)
	posn := snapshot.Position(nil)
	s.Physics.SetPosition(cp.Vector{float64(posn.X()), float64(posn.Y())})
	s.Physics.SetAngle(float64(snapshot.Rotation()))
	s.Armed = snapshot.Armed() > 0
	s.Thrusting = snapshot.Thrusting() > 0
	s.Health = int(snapshot.Health())
}

type Entity interface{}

type System struct {
	manager *entity.Manager
	window  *pixelgl.Window
	stateMu sync.RWMutex
	state   map[entity.ID]Entity
	inputs  *queue.RingBuffer

	renderer *rendering.GameRenderer
	world    *world.World
}

func New(manager *entity.Manager, window *pixelgl.Window, host string) *System {
	system := &System{
		manager: manager,
		window:  window,
		inputs:  queue.NewRingBuffer(4),
		state:   make(map[entity.ID]Entity),
	}
	go system.loop(host)
	return system
}

func (s *System) loop(host string) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		message := fbs.GetRootAsMessage(data, 0)
		packetTable := new(flatbuffers.Table)
		if !message.Packet(packetTable) {
			continue
		}
		switch message.PacketType() {
		case fbs.PacketServer:
			server := new(fbs.Server)
			server.Init(packetTable.Bytes, packetTable.Pos)
			// todo initialize other systems like physics

		case fbs.PacketSpawn:
			spawn := new(fbs.Spawn)
			spawn.Init(packetTable.Bytes, packetTable.Pos)
			if s.renderer != nil {
				panic("networking: received spawn packet while alive")
			}
			s.renderer = rendering.NewGameRenderer(s.window, spawn.Id())
		case fbs.PacketPerception:
			perception := new(fbs.Perception)
			perception.Init(packetTable.Bytes, packetTable.Pos)
			s.stateMu.Lock()
			s.world.Do(func(space *cp.Space) {
				for i := 0; i < perception.EntitiesLength(); i++ {
					entitySnap := new(fbs.Entity)
					if ok := perception.Entities(entitySnap, i); ok {
						snapshotTable := new(flatbuffers.Table)
						if ok := entitySnap.Snapshot(snapshotTable); ok {
							switch entitySnap.SnapshotType() {
							case fbs.SnapshotShip:
								ship := new(fbs.Ship)
								ship.Init(snapshotTable.Bytes, snapshotTable.Pos)
								e, ok := s.state[ship.Id()]
								if ok {
									if es, ok := e.(*shipEntity); ok {
										es.Health = int(ship.Health())
										es.Thrusting = ship.Thrusting() > 0
										es.Armed = ship.Armed() > 0
										p := ship.Position(nil)
										es.Physics.SetPosition(cp.Vector{X: float64(p.X()), Y: float64(p.Y())})
									}
								} else {
									s.state[ship.Id()] = &shipEntity{
										ID:      ship.Id(),
										Name:    string(ship.Name()),
										Physics: physics.NewShip(space, ship.Id()),
										Rendering: rendering.Component{
											//	Sprite:pixel.NewSprite() //todo finish rendering
										},
										Armed:     ship.Armed() > 0,
										Thrusting: ship.Thrusting() > 0,
										Health:    int(ship.Health()),
									}
								}
							}
						}
					}
				}
			})
			s.stateMu.Unlock()
		}
	}
}

func (s *System) Update(delta float64) {
	// todo send inputs
}

func (s *System) Remove(entity entity.ID) {
}
