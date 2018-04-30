package networking

import (
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/server/networking/fbs"
)

func messageC(builder *flatbuffers.Builder, packet flatbuffers.UOffsetT, packetType byte) []byte {
	fbs.MessageStart(builder)
	fbs.MessageAddPacket(builder, packet)
	fbs.MessageAddPacketType(builder, packetType)
	m := fbs.MessageEnd(builder)
	builder.Finish(m)
	var bytes = make([]byte, len(builder.FinishedBytes()))
	copy(bytes, builder.FinishedBytes())
	return bytes
}
