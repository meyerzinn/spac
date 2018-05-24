package ship

import (
	"github.com/20zinnm/spac/server/movement"
	"github.com/20zinnm/spac/server/shooting"
)

type Controls struct {
	Movement movement.Controls
	Shooting shooting.Controls
}
