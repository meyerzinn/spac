package net

import (
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/common/net/fbs"
)

func Message(builder *flatbuffers.Builder, packet flatbuffers.UOffsetT, packetType byte) []byte {
	fbs.MessageStart(builder)
	fbs.MessageAddPacket(builder, packet)
	fbs.MessageAddPacketType(builder, packetType)
	m := fbs.MessageEnd(builder)
	builder.Finish(m)
	return builder.FinishedBytes()
}

