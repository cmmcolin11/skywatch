package msgpack

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
)

const (
	bytesAllocLimit = 1e6
	sliceAllocLimit = 1e4
	maxMapSize      = 1e6
)

type bufReader interface {
	io.Reader
	io.ByteScanner
}

type Decoder struct {
	r          io.Reader
	s          io.ByteScanner
	buf        []byte
	rec        []byte
	mapDecoder func(*Decoder) (interface{}, error)
}

var decPool = sync.Pool{
	New: func() interface{} {
		return NewDecoder(nil)
	},
}

func NewDecoder(r io.Reader) *Decoder {
	d := new(Decoder)
	d.Reset(r)
	return d
}

func GetDecoder() *Decoder {
	return decPool.Get().(*Decoder)
}

func PutDecoder(dec *Decoder) {
	dec.r = nil
	dec.s = nil
	decPool.Put(dec)
}

func Unmarshal(data []byte, v interface{}) error {
	dec := GetDecoder()

	dec.Reset(bytes.NewReader(data))
	if err := dec.Decode(v); err != nil {
		return err
	}

	PutDecoder(dec)
	return nil
}

func (d *Decoder) Reset(r io.Reader) {
	if br, ok := r.(bufReader); ok {
		d.r = br
		d.s = br
	} else {
		br := bufio.NewReader(r)
		d.r = br
		d.s = br
	}
	// d.flags = 0
	// d.structTag = ""
	// d.mapDecoder = nil
	// d.dict = nil
}

func (d *Decoder) Decode(v interface{}) error {
	switch v := v.(type) {
	case *map[string]interface{}:
		m, err := d.DecodeMap()
		if err != nil {
			return err
		}
		*v = m
		return nil
	default:
		return errors.New(fmt.Sprintf("Not Support: %T", v))
	}
}

//--------------------------------------------------

func (d *Decoder) DecodeMap() (map[string]interface{}, error) {
	n, err := d.DecodeMapLen()
	if err != nil {
		return nil, err
	}

	if n == -1 {
		return nil, nil
	}

	m := make(map[string]interface{}, min(n, maxMapSize))

	for i := 0; i < n; i++ {
		mk, err := d.DecodeString()
		if err != nil {
			return nil, err
		}
		mv, err := d.DecodeInterface()
		if err != nil {
			return nil, err
		}
		m[mk] = mv
	}

	return m, nil
}

func (d *Decoder) DecodeMapLen() (int, error) {
	c, err := d.readCode()
	if err != nil {
		return 0, err
	}
	return d.mapLen(c)
}

func (d *Decoder) readCode() (byte, error) {
	c, err := d.s.ReadByte()
	if err != nil {
		fmt.Println("readCode:", err)
		return 0, err
	}
	if d.rec != nil {
		d.rec = append(d.rec, c)
	}
	return c, nil
}

func (d *Decoder) mapLen(c byte) (int, error) {
	if c == Nil {
		return -1, nil
	}
	if c >= FixedMapLow && c <= FixedMapHigh {
		return int(c & FixedMapMask), nil
	}
	if c == Map16 {
		size, err := d.uint16()
		return int(size), err
	}
	if c == Map32 {
		size, err := d.uint32()
		return int(size), err
	}
	return 0, unexpectedCodeError{code: c, hint: "map length"}
}

//--------------------------------------------------

func (d *Decoder) DecodeString() (string, error) {
	c, err := d.readCode()
	if err != nil {
		fmt.Println("DecodeString:", err)
		return "", err
	}
	return d.string(c)
}

func (d *Decoder) string(c byte) (string, error) {
	n, err := d.bytesLen(c)
	if err != nil {
		return "", err
	}
	return d.stringWithLen(n)
}

func (d *Decoder) bytesLen(c byte) (int, error) {
	if c == Nil {
		fmt.Println("bytesLen:", c)
		return -1, nil
	}

	if IsFixedString(c) {
		return int(c & FixedStrMask), nil
	}

	switch c {
	case Str8, Bin8:
		n, err := d.uint8()
		return int(n), err
	case Str16, Bin16:
		n, err := d.uint16()
		return int(n), err
	case Str32, Bin32:
		n, err := d.uint32()
		return int(n), err
	}

	return 0, fmt.Errorf("msgpack: invalid code=%x decoding string/bytes length", c)
}

func (d *Decoder) stringWithLen(n int) (string, error) {
	if n <= 0 {
		return "", nil
	}
	b, err := d.readN(n)
	return string(b), err
}

//--------------------------------------------------

func (d *Decoder) DecodeInterface() (interface{}, error) {
	c, err := d.readCode()
	if err != nil {
		return nil, err
	}

	if IsFixedNum(c) {
		return int8(c), nil
	}
	if IsFixedMap(c) {
		err = d.s.UnreadByte()
		if err != nil {
			return nil, err
		}
		return d.decodeMapDefault()
	}
	if IsFixedArray(c) {
		return d.decodeSlice(c)
	}
	if IsFixedString(c) {
		return d.string(c)
	}

	switch c {
	case Nil:
		return nil, nil
	case False, True:
		return d.bool(c)
	case Double:
		return d.float64(c)
	case Str8, Str16, Str32:
		return d.string(c)
	case Map16, Map32:
		err = d.s.UnreadByte()
		if err != nil {
			return nil, err
		}
		return d.decodeMapDefault()
	default:
		fmt.Printf("Not support 0x%s\n", hex.EncodeToString([]byte{c}))
	}

	return 0, fmt.Errorf("msgpack: unknown code %x decoding interface{}", c)
}

func (d *Decoder) int(c byte) (int64, error) {
	if c == Nil {
		return 0, nil
	}
	if IsFixedNum(c) {
		return int64(int8(c)), nil
	}
	switch c {
	case Uint8:
		n, err := d.uint8()
		return int64(n), err
	case Int8:
		n, err := d.uint8()
		return int64(int8(n)), err
	case Uint16:
		n, err := d.uint16()
		return int64(n), err
	case Int16:
		n, err := d.uint16()
		return int64(int16(n)), err
	case Uint32:
		n, err := d.uint32()
		return int64(n), err
	case Int32:
		n, err := d.uint32()
		return int64(int32(n)), err
	case Uint64, Int64:
		n, err := d.uint64()
		return int64(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code=%x decoding int64", c)
}

func (d *Decoder) int8() (int8, error) {
	n, err := d.uint8()
	return int8(n), err
}

func (d *Decoder) float32(c byte) (float32, error) {
	if c == Float {
		n, err := d.uint32()
		if err != nil {
			return 0, err
		}
		return math.Float32frombits(n), nil
	}

	n, err := d.int(c)
	if err != nil {
		return 0, fmt.Errorf("msgpack: invalid code=%x decoding float32", c)
	}
	return float32(n), nil
}

func (d *Decoder) float64(c byte) (float64, error) {
	switch c {
	case Float:
		n, err := d.float32(c)
		if err != nil {
			return 0, err
		}
		return float64(n), nil
	case Double:
		n, err := d.uint64()
		if err != nil {
			return 0, err
		}
		return math.Float64frombits(n), nil
	}

	n, err := d.int(c)
	if err != nil {
		return 0, fmt.Errorf("msgpack: invalid code=%x decoding float32", c)
	}
	return float64(n), nil
}

func (d *Decoder) decodeMapDefault() (interface{}, error) {
	if d.mapDecoder != nil {
		return d.mapDecoder(d)
	}
	return d.DecodeMap()
}

func (d *Decoder) decodeSlice(c byte) ([]interface{}, error) {
	n, err := d.arrayLen(c)
	if err != nil {
		return nil, err
	}
	if n == -1 {
		return nil, nil
	}

	s := make([]interface{}, 0, min(n, sliceAllocLimit))
	for i := 0; i < n; i++ {
		v, err := d.DecodeInterface()
		if err != nil {
			return nil, err
		}
		s = append(s, v)
	}

	return s, nil
}

func (d *Decoder) arrayLen(c byte) (int, error) {
	if c == Nil {
		return -1, nil
	} else if c >= FixedArrayLow && c <= FixedArrayHigh {
		return int(c & FixedArrayMask), nil
	}
	switch c {
	case Array16:
		n, err := d.uint16()
		return int(n), err
	case Array32:
		n, err := d.uint32()
		return int(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code=%x decoding array length", c)
}

//--------------------------------------------------

func (d *Decoder) DecodeBool() (bool, error) {
	c, err := d.readCode()
	if err != nil {
		return false, err
	}
	return d.bool(c)
}

func (d *Decoder) bool(c byte) (bool, error) {
	if c == Nil {
		return false, nil
	}
	if c == False {
		return false, nil
	}
	if c == True {
		return true, nil
	}
	return false, fmt.Errorf("msgpack: invalid code=%x decoding bool", c)
}

//--------------------------------------------------

func (d *Decoder) uint8() (uint8, error) {
	c, err := d.readCode()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d *Decoder) uint16() (uint16, error) {
	b, err := d.readN(2)
	if err != nil {
		return 0, err
	}
	return (uint16(b[0]) << 8) | uint16(b[1]), nil
}

func (d *Decoder) uint32() (uint32, error) {
	b, err := d.readN(4)
	if err != nil {
		return 0, err
	}
	n := (uint32(b[0]) << 24) |
		(uint32(b[1]) << 16) |
		(uint32(b[2]) << 8) |
		uint32(b[3])
	return n, nil
}

func (d *Decoder) uint64() (uint64, error) {
	b, err := d.readN(8)
	if err != nil {
		return 0, err
	}
	n := (uint64(b[0]) << 56) |
		(uint64(b[1]) << 48) |
		(uint64(b[2]) << 40) |
		(uint64(b[3]) << 32) |
		(uint64(b[4]) << 24) |
		(uint64(b[5]) << 16) |
		(uint64(b[6]) << 8) |
		uint64(b[7])
	return n, nil
}

func (d *Decoder) readN(n int) ([]byte, error) {
	var err error
	d.buf, err = readN(d.r, d.buf, n)
	if err != nil {
		return nil, err
	}
	if d.rec != nil {
		d.rec = append(d.rec, d.buf...)
	}
	return d.buf, nil
}
