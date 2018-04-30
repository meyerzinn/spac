package despawning

type Component struct {
	// TTL is the number of ticks before an entity de-spawns.
	TTL   uint
	alive uint
}
