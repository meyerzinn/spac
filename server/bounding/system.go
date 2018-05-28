package bounding

//
//type boundingEntity struct {
//	Physics world.Component
//	Health  *damaging.Component
//	Decay   float64
//}
//
//type System struct {
//	world      *world.World
//	radius     float64
//	entitiesMu sync.RWMutex
//	entities   map[entity.ID]boundingEntity
//}
//
//func New(world *world.World, radius float64) *System {
//	return &System{
//		world:    world,
//		radius:   radius,
//		entities: make(map[entity.ID]boundingEntity),
//	}
//}
//
//func (s *System) Add(id entity.ID, physics world.Component, health *damaging.Component, decay float64) {
//	s.entitiesMu.Lock()
//	s.entities[id] = boundingEntity{
//		Physics: physics,
//		Health:  health,
//		Decay:   decay,
//	}
//	s.entitiesMu.Unlock()
//}
//
//func (s *System) Update(delta float64) {
//	s.entitiesMu.RLock()
//	defer s.entitiesMu.RUnlock()
//	for _, entity := range s.entities {
//		if !entity.Physics.Position().Near(cp.Vector{}, s.radius) {
//			entity.Health.Value -= entity.Decay * entity.Health.Max
//		}
//	}
//}
//
//func (s *System) Remove(entity entity.ID) {
//	panic("implement me")
//}
