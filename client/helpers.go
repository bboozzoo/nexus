package client

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gammazero/nexus/v3/wamp"
)

var (
	errArgumentIsNil = errors.New("argument is nil")
)

func unpackSimple(src, dst reflect.Value) error {
	fmt.Printf("-- unpack\n")
	fmt.Printf("dst: %+v %v\n", dst, dst.Type())
	fmt.Printf("src: %+v %v\n", src, src.Type())
	fmt.Printf("assignable? %v\n", src.Type().AssignableTo(dst.Type()))
	fmt.Printf("convertible? %v\n", src.Type().ConvertibleTo(dst.Type()))

	if src.Kind() == reflect.Interface {
		fmt.Printf("src is interface{}\n")
		return unpackSimple(src.Elem(), dst)
	}

	if !src.Type().ConvertibleTo(dst.Type()) {
		return fmt.Errorf("cannot convert %v to %v", src.Type(), dst.Type())
	}
	dst.Set(src.Convert(dst.Type()))
	return nil
}

func unpackSlice(src, dst reflect.Value) error {
	fmt.Printf("-- unpack slice\n")
	if dst.Len() < src.Len() {
		dst.Set(reflect.MakeSlice(dst.Type(), src.Len(), src.Cap()))
	}
	for i := 0; i < src.Len(); i++ {
		if err := unpackSimple(src.Index(i), dst.Index(i)); err != nil {
			return fmt.Errorf("cannot unpack slice element %v: %v", i, err)
		}
	}
	return nil
}

func unpackMap(src, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.Struct:
		return unpackMapToStruct(src, dst)
	case reflect.Map:
		return unpackMapToMap(src, dst)
	default:
		return fmt.Errorf("cannot unpack map to %v", dst.Kind())
	}
}

func unpackMapToMap(src, dst reflect.Value) error {
	fmt.Printf("-- unpack map\n")
	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}
	for _, sk := range src.MapKeys() {
		dk := sk.Convert(dst.Type().Key())
		dv := reflect.New(dst.Type().Elem()).Elem()
		if err := unpackSimple(src.MapIndex(sk), dv); err != nil {
			return fmt.Errorf("cannot unpack map element %v: %v", sk, err)
		}
		dst.SetMapIndex(dk, dv)
	}
	return nil
}

func unpackMapToStruct(src, dst reflect.Value) error {
	fmt.Printf("-- unpack map to struct\n")
	dtype := dst.Type() // struct's type
	for i := 0; i < dst.NumField(); i++ {
		fv := dst.Field(i)
		ft := dtype.Field(i) // field's type
		fmt.Printf("--- to %v | %v\n", ft, fv)
		var optional bool
		var keyName string
		if wt := ft.Tag.Get("wamp"); wt != "" {
			// wamp:"-"
			if wt == "-" {
				continue
			}
			// wamp:"name,<opts>"
			idx := strings.LastIndex(wt, ",")
			var opts string
			if idx != -1 {
				opts = wt[idx+1:]
				if opts == "optional" {
					optional = true
				}
				keyName = wt[:idx]
			} else {
				// find key by tag value
				keyName = wt
			}
		}
		if keyName == "" {
			// find key by field name
			keyName = ft.Name
		}
		mk := reflect.ValueOf(keyName)
		fmt.Printf("mk %v\n", mk)
		srcv := src.MapIndex(mk)
		fmt.Printf("srcv: %v", srcv)
		if !srcv.IsValid() {
			if !optional {
				return fmt.Errorf("required field %q not found", keyName)
			}
			// field not found in map
			continue
		}
		if err := unpackSimple(srcv, fv); err != nil {
			return fmt.Errorf("cannot unpack map field: %v", err)
		}
	}
	return nil
}

func unpackOneFromTo(src, dst reflect.Value) error {
	if dst.Kind() != reflect.Pointer || dst.IsNil() {
		return errArgumentIsNil
	}
	switch src.Kind() {
	case reflect.Slice:
		return unpackSlice(src, dst.Elem())
	case reflect.Map:
		return unpackMap(src, dst.Elem())
	default:
		return unpackSimple(src, dst.Elem())
	}
}

// Unpack is a convenience helper for unpacking arguments of an RPC call or an
// Event.
func Unpack(args wamp.List, out ...interface{}) error {
	if len(args) != len(out) {
		return fmt.Errorf("unexpected number of arguments")
	}
	for i := range args {
		err := unpackOneFromTo(reflect.ValueOf(args[i]),
			reflect.ValueOf(out[i]))
		if err != nil {
			return fmt.Errorf("cannot unpack argument %v: %w", i, err)
		}
	}
	return nil
}

// UnpackDict is a convenience helper for unpacking keyword arguments of an RPC
// call or an event into a structure. Keys are unpacked to struct fields by the
// field's name or the field `wamp` tag. Example:
//
// Given a struct:
// type s struct {
//     Foo int
//     Bar int `wamp:"bar"`
// }
//
// a WAMP keyword arguments
//
// wamp.Dict{
//     "Foo": 123
//     "bar": 345,
// }
//
// will match:
// "Foo" -> s.Foo
// "bar" -> s.bar
// TODO required fields by default, add ,optional tag
func UnpackDict(dict wamp.Dict, out interface{}) error {
	return unpackOneFromTo(reflect.ValueOf(dict), reflect.ValueOf(out))
}

// ErrWithMessage returns an InvokeResult filled with given error URI and a
// message passed as the first argument in the argument list.
func ErrWithMessage(err wamp.URI, message string) InvokeResult {
	return InvokeResult{
		Err:  err,
		Args: wamp.List{message},
	}
}
