package msgpack

import (
	"fmt"
	"reflect"
	"sync"
)

type (
	encoderFunc func(*Encoder, reflect.Value) error
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()

var (
	typeEncMap    sync.Map
	valueEncoders []encoderFunc
)

func init() {
	valueEncoders = []encoderFunc{
		reflect.Bool:      encodeBoolValue,
		reflect.Float64:   encodeFloat64Value,
		reflect.Int:       encodeIntValue,
		reflect.Interface: encodeInterfaceValue,
		reflect.Map:       encodeMapValue,
		reflect.Slice:     encodeSliceValue,
		reflect.String:    encodeStringValue,
	}
}

func getEncoder(typ reflect.Type) encoderFunc {
	if v, ok := typeEncMap.Load(typ); ok {
		return v.(encoderFunc)
	}

	if typ == errorType {
		return encodeErrorValue
	}

	kind := typ.Kind()
	// en:fmt.Println("kind: ", kind)

	fn := valueEncoders[kind]
	if fn == nil {
		return encodeNotFound
	}

	typeEncMap.Store(typ, fn)
	return fn
}

func encodeBoolValue(e *Encoder, v reflect.Value) error {
	return e.EncodeBool(v.Bool())
}

func encodeFloat64Value(e *Encoder, v reflect.Value) error {
	return e.EncodeFloat64(v.Float())
}

func encodeIntValue(e *Encoder, v reflect.Value) error {
	return e.EncodeInt(v.Int())
}

func encodeInterfaceValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}
	return e.EncodeValue(v.Elem())
}

func encodeMapValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	if err := e.EncodeMapLen(v.Len()); err != nil {
		return err
	}

	iter := v.MapRange()
	for iter.Next() {
		if err := e.EncodeValue(iter.Key()); err != nil {
			return err
		}
		if err := e.EncodeValue(iter.Value()); err != nil {
			return err
		}
	}

	return nil
}

func encodeStringValue(e *Encoder, v reflect.Value) error {
	return e.EncodeString(v.String())
}

func encodeSliceValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	l := v.Len()
	if err := e.EncodeArrayLen(l); err != nil {
		return err
	}
	for i := 0; i < l; i++ {
		if err := e.EncodeValue(v.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func encodeErrorValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}
	return e.EncodeString(v.Interface().(error).Error())
}

func encodeNotFound(e *Encoder, v reflect.Value) error {
	return fmt.Errorf("encode map key(%s) not found", v.Type().Kind().String())
}
