package reflection

import "reflect"

// IsZero returns true if the value v is the zero value for its type.
// Works for basic types (int, string, bool, etc.) and structs.
func IsZero(v any) bool {
	return reflect.ValueOf(v).IsZero()
}

// IsNil returns true if v is nil, or if v is a typed nil (ptr, slice, map, chan, func, interface).
// For value types (struct, int, string, etc.), it always returns false.
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return rv.IsNil()
	default:
		return false
	}
}
