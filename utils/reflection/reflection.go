package reflection

import "reflect"

// IsZero returns true if the value v is the zero value for its type.
// Works for basic types (int, string, bool, etc.) and structs.
func IsZero(v any) bool {
	return reflect.ValueOf(v).IsZero()
}
