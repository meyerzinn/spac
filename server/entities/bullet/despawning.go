package bullet

import "github.com/20zinnm/spac/server/despawning"

func Despawning(ttl uint) despawning.Component {
	return despawning.Component{
		TTL: ttl,
	}
}
