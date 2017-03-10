package encoding

import (
	"fmt"
	"io"
	"reflect"

	"github.com/djwackey/amf0"
)

// Unmarshaler fills a struct with data from an AMF stream.
type Unmarshaler struct {
	r io.Reader

	decode amf0.Decoder
}

// Unmarshal creates a new Unmarshaler and directly unmarshals onto the given
// type from the given io.Reader, returning any errors that it encounters along
// the way.
func Unmarshal(r io.Reader, v interface{}) error {
	return NewUnmarshaler(r).Unmarshal(v)
}

// NewUnmarshaler creates a new Unmarshaler initialized with the default AMF
// decoder, and the given io.Reader.
func NewUnmarshaler(r io.Reader) *Unmarshaler {
	return &Unmarshaler{
		r:      r,
		decode: amf0.Decode,
	}
}

// Unmarshal fills each field in the givne interface{} with the AMF data on the
// stream in-order.
//
// If a value of amf0.Null or amf0.Undefined is read, then the value will be
// skipped.
//
// If a pointer field type is reached, then the value will be reduced to match
// that same pointer type.
func (u *Unmarshaler) Unmarshal(dest interface{}) error {
	v := reflect.ValueOf(dest).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		next, err := u.decode(u.r)
		if err != nil {
			return err
		}

		if u.isBodyless(next) {
			if !u.canAssignNil(field) {
				return fmt.Errorf(
					"amf0: unable to assign nil to type %T",
					field.Interface())
			}

			continue
		}

		var val reflect.Value
		if u.canAssignNil(field) {
			val = reflect.ValueOf(next)
		} else {
			val = reflect.ValueOf(next).Elem()
		}

		field.Set(val.Convert(field.Type()))
	}

	return nil
}

func (u *Unmarshaler) canAssignNil(f reflect.Value) bool {
	kind := f.Kind()

	return kind == reflect.Array ||
		kind == reflect.Chan ||
		kind == reflect.Func ||
		kind == reflect.Interface ||
		kind == reflect.Map ||
		kind == reflect.Ptr ||
		kind == reflect.Slice
}

// isBodyless returns a bool representing whether or not the given amf0.Type is
// bodyless or not.
func (u *Unmarshaler) isBodyless(t amf0.AmfType) bool {
	if v, isBodyless := t.(amf0.Bodyless); isBodyless {
		return v.IsBodyless()
	}

	return false
}
