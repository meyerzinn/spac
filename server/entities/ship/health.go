package ship

import "github.com/20zinnm/spac/server/health"

const StartingHealth health.Component = 100

func Health() *health.Component {
	var health = health.Component(StartingHealth)
	return &health
}
