package p2p

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *DiscReason) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 uint
		zb0001, err = dc.ReadUint()
		if err != nil {
			return
		}
		(*z) = DiscReason(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z DiscReason) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteUint(uint(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z DiscReason) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendUint(o, uint(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *DiscReason) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 uint
		zb0001, bts, err = msgp.ReadUintBytes(bts)
		if err != nil {
			return
		}
		(*z) = DiscReason(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z DiscReason) Msgsize() (s int) {
	s = msgp.UintSize
	return
}
