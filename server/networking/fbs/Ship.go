// automatically generated by the FlatBuffers compiler, do not modify

package fbs

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Ship struct {
	_tab flatbuffers.Table
}

func GetRootAsShip(buf []byte, offset flatbuffers.UOffsetT) *Ship {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Ship{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Ship) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Ship) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Ship) Id() uint64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetUint64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Ship) MutateId(n uint64) bool {
	return rcv._tab.MutateUint64Slot(4, n)
}

func (rcv *Ship) Position(obj *Point) *Point {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		x := o + rcv._tab.Pos
		if obj == nil {
			obj = new(Point)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *Ship) Rotation() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Ship) MutateRotation(n float32) bool {
	return rcv._tab.MutateFloat32Slot(8, n)
}

func (rcv *Ship) Thrusting() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Ship) MutateThrusting(n byte) bool {
	return rcv._tab.MutateByteSlot(10, n)
}

func (rcv *Ship) Armed() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Ship) MutateArmed(n byte) bool {
	return rcv._tab.MutateByteSlot(12, n)
}

func (rcv *Ship) Health() int16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Ship) MutateHealth(n int16) bool {
	return rcv._tab.MutateInt16Slot(14, n)
}

func (rcv *Ship) Name() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func ShipStart(builder *flatbuffers.Builder) {
	builder.StartObject(7)
}
func ShipAddId(builder *flatbuffers.Builder, id uint64) {
	builder.PrependUint64Slot(0, id, 0)
}
func ShipAddPosition(builder *flatbuffers.Builder, position flatbuffers.UOffsetT) {
	builder.PrependStructSlot(1, flatbuffers.UOffsetT(position), 0)
}
func ShipAddRotation(builder *flatbuffers.Builder, rotation float32) {
	builder.PrependFloat32Slot(2, rotation, 0.0)
}
func ShipAddThrusting(builder *flatbuffers.Builder, thrusting byte) {
	builder.PrependByteSlot(3, thrusting, 0)
}
func ShipAddArmed(builder *flatbuffers.Builder, armed byte) {
	builder.PrependByteSlot(4, armed, 0)
}
func ShipAddHealth(builder *flatbuffers.Builder, health int16) {
	builder.PrependInt16Slot(5, health, 0)
}
func ShipAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(6, flatbuffers.UOffsetT(name), 0)
}
func ShipEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}