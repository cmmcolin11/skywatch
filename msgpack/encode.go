package msgpack

import (
	"bytes"
	"io"
	"math"
	"reflect"
	"sync"
)

type writer interface {
	io.Writer
	WriteByte(byte) error
}

type Encoder struct {
	w   writer
	buf []byte
}

var encPool = sync.Pool{
	New: func() interface{} {
		return NewEncoder(nil)
	},
}

func NewEncoder(w io.Writer) *Encoder {
	e := &Encoder{
		buf: make([]byte, 9),
	}
	e.Reset(w)
	return e
}

func Marshal(v interface{}) ([]byte, error) {
	enc := GetEncoder()

	var buf bytes.Buffer
	enc.Reset(&buf)

	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	b := buf.Bytes()
	return b, nil

}

func GetEncoder() *Encoder {
	return encPool.Get().(*Encoder)
}

func (e *Encoder) Encode(v interface{}) error {
	return e.EncodeValue(reflect.ValueOf(v))
}

func (e *Encoder) EncodeValue(v reflect.Value) error {
	fn := getEncoder(v.Type())
	return fn(e, v)
}

func (e *Encoder) EncodeNil() error {
	return e.writeCode(Nil)
}

func (e *Encoder) EncodeBool(value bool) error {
	if value {
		return e.writeCode(True)
	}
	return e.writeCode(False)
}

func (e *Encoder) EncodeMapLen(l int) error {
	if l < 16 {
		return e.writeCode(FixedMapLow | byte(l))
	}
	if l <= math.MaxUint16 {
		return e.write2(Map16, uint16(l))
	}
	return e.write4(Map32, uint32(l))
}

func (e *Encoder) EncodeString(v string) error {
	if err := e.encodeStringLen(len(v)); err != nil {
		return err
	}
	return e.writeString(v)
}

func (e *Encoder) EncodeInt(n int64) error {
	if n >= 0 {
		return e.EncodeUint(uint64(n))
	}
	if n >= int64(int8(NegFixedNumLow)) {
		return e.w.WriteByte(byte(n))
	}
	if n >= math.MinInt8 {
		return e.EncodeInt8(int8(n))
	}
	if n >= math.MinInt16 {
		return e.EncodeInt16(int16(n))
	}
	if n >= math.MinInt32 {
		return e.EncodeInt32(int32(n))
	}
	return e.EncodeInt64(n)
}

func (e *Encoder) EncodeInt8(n int8) error {
	return e.write1(Int8, uint8(n))
}

func (e *Encoder) EncodeInt16(n int16) error {
	return e.write2(Int16, uint16(n))
}

func (e *Encoder) EncodeInt32(n int32) error {
	return e.write4(Int32, uint32(n))
}

func (e *Encoder) EncodeInt64(n int64) error {
	return e.write8(Int64, uint64(n))
}

func (e *Encoder) EncodeUint(n uint64) error {
	if n <= math.MaxInt8 {
		return e.w.WriteByte(byte(n))
	}
	if n <= math.MaxUint8 {
		return e.EncodeUint8(uint8(n))
	}
	if n <= math.MaxUint16 {
		return e.EncodeUint16(uint16(n))
	}
	if n <= math.MaxUint32 {
		return e.EncodeUint32(uint32(n))
	}
	return e.EncodeUint64(n)
}

func (e *Encoder) EncodeUint8(n uint8) error {
	return e.write1(Uint8, n)
}

func (e *Encoder) EncodeUint16(n uint16) error {
	return e.write2(Uint16, n)
}

func (e *Encoder) EncodeUint32(n uint32) error {
	return e.write4(Uint32, n)
}

func (e *Encoder) EncodeUint64(n uint64) error {
	return e.write8(Uint64, n)
}

func (e *Encoder) EncodeFloat64(n float64) error {
	return e.write8(Double, math.Float64bits(n))
}

func (e *Encoder) EncodeArrayLen(l int) error {
	if l < 16 {
		return e.writeCode(FixedArrayLow | byte(l))
	}
	if l <= math.MaxUint16 {
		return e.write2(Array16, uint16(l))
	}
	return e.write4(Array32, uint32(l))
}

func (e *Encoder) write(b []byte) error {
	_, err := e.w.Write(b)
	return err
}

func (e *Encoder) writeCode(c byte) error {
	return e.w.WriteByte(c)
}

func (e *Encoder) write1(code byte, n uint8) error {
	e.buf = e.buf[:2]
	e.buf[0] = code
	e.buf[1] = n
	return e.write(e.buf)
}

func (e *Encoder) write2(code byte, n uint16) error {
	e.buf = e.buf[:3]
	e.buf[0] = code
	e.buf[1] = byte(n >> 8)
	e.buf[2] = byte(n)
	return e.write(e.buf)
}

func (e *Encoder) write4(code byte, n uint32) error {
	e.buf = e.buf[:5]
	e.buf[0] = code
	e.buf[1] = byte(n >> 24)
	e.buf[2] = byte(n >> 16)
	e.buf[3] = byte(n >> 8)
	e.buf[4] = byte(n)
	return e.write(e.buf)
}

func (e *Encoder) write8(code byte, n uint64) error {
	e.buf = e.buf[:9]
	e.buf[0] = code
	e.buf[1] = byte(n >> 56)
	e.buf[2] = byte(n >> 48)
	e.buf[3] = byte(n >> 40)
	e.buf[4] = byte(n >> 32)
	e.buf[5] = byte(n >> 24)
	e.buf[6] = byte(n >> 16)
	e.buf[7] = byte(n >> 8)
	e.buf[8] = byte(n)
	return e.write(e.buf)
}

func (e *Encoder) Reset(w io.Writer) {
	if bw, ok := w.(writer); ok {
		e.w = bw
	} else {
		e.w = newByteWriter(w)
	}
}

func (e *Encoder) Writer() io.Writer {
	return e.w
}

func (e *Encoder) encodeStringLen(l int) error {
	if l < 32 {
		return e.writeCode(FixedStrLow | byte(l))
	}
	if l < 256 {
		return e.write1(Str8, uint8(l))
	}
	if l <= math.MaxUint16 {
		return e.write2(Str16, uint16(l))
	}
	return e.write4(Str32, uint32(l))
}

func (e *Encoder) writeString(s string) error {
	_, err := e.w.Write(stringToBytes(s))
	return err
}

func stringToBytes(s string) []byte {
	return []byte(s)
}

//--------------------------------------------------

type byteWriter struct {
	io.Writer
}

func (bw byteWriter) WriteByte(c byte) error {
	_, err := bw.Write([]byte{c})
	return err
}

func newByteWriter(w io.Writer) byteWriter {
	return byteWriter{
		Writer: w,
	}
}
