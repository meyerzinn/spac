package net

import (
	"github.com/google/flatbuffers/go"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/20zinnm/spac/common/net/upstream"
)

func MessageDown(builder *flatbuffers.Builder, packetType byte, packet flatbuffers.UOffsetT ) []byte {
	downstream.MessageStart(builder)
	downstream.MessageAddPacket(builder, packet)
	downstream.MessageAddPacketType(builder, packetType)
	m := downstream.MessageEnd(builder)
	builder.Finish(m)
	return builder.FinishedBytes()
}

func MessageUp(builder *flatbuffers.Builder, packetType byte, packet flatbuffers.UOffsetT) []byte {
	upstream.MessageStart(builder)
	upstream.MessageAddPacket(builder, packet)
	upstream.MessageAddPacketType(builder, packetType)
	m := upstream.MessageEnd(builder)
	builder.Finish(m)
	return builder.FinishedBytes()
}
