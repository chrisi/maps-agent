package camtac

import (
	"encoding/binary"
	"math"
)

type Cursor struct {
	data []byte
	pos  int
}

func (c *Cursor) Int8() int {
	return int(c.Uint8())
}

func (c *Cursor) Int16() int16 {
	return int16(c.Uint16())
}

func (c *Cursor) Int32() int32 {
	return int32(c.Uint32())
}

func (c *Cursor) Int64() int {
	return int(c.Uint64())
}

func NewCursor(data []byte) *Cursor {
	return &Cursor{data: data}
}

func (c *Cursor) Uint8() uint8 {
	v := c.data[c.pos]
	c.pos++
	return v
}

func (c *Cursor) Uint16() uint16 {
	v := binary.LittleEndian.Uint16(c.data[c.pos : c.pos+2])
	c.pos += 2
	return v
}

func (c *Cursor) Uint32() uint32 {
	v := binary.LittleEndian.Uint32(c.data[c.pos : c.pos+4])
	c.pos += 4
	return v
}

func (c *Cursor) Uint64() uint64 {
	v := binary.LittleEndian.Uint64(c.data[c.pos : c.pos+8])
	c.pos += 8
	return v
}

func (c *Cursor) Float32() float32 {
	bits := c.Uint32()
	return math.Float32frombits(bits)
}

func (c *Cursor) Float64() float64 {
	bits := c.Uint64()
	return math.Float64frombits(bits)
}

func (c *Cursor) Byte() byte {
	v := c.data[c.pos]
	c.pos++
	return v
}

func (c *Cursor) Bytes(n int) []byte {
	v := c.data[c.pos : c.pos+n]
	c.pos += n
	return v
}
