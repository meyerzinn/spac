package ship

import (
	"github.com/20zinnm/spac/server/shooting"
)

func Shooting() *shooting.Component {
	return &shooting.Component{
		Cooldown:       20,
		BulletForce:    500,
		BulletLifetime: 100,
	}
}
