// builders implements a pool of flatbuffers Builders and prevents data leaks in pool members.
package builders

import (
	"sync"
	"github.com/google/flatbuffers/go"
)

const (
	InitialSize = 64
)

var builderPool = sync.Pool{
	New: func() interface{} {
		return flatbuffers.NewBuilder(InitialSize)
	},
}

func Get() *flatbuffers.Builder {
	b := builderPool.Get().(*flatbuffers.Builder)
	b.Reset()
	return b
}

func Put(b *flatbuffers.Builder) {
	b.Reset()
	builderPool.Put(b)
}
