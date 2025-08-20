package generic

// SafePtr safely applies a transformation function to a pointer,
// returning a new pointer to the result if the source is not nil.
//
// Parameters:
//   - src: a pointer to the input value of type T
//   - get: a function that takes a *T and returns a value of type R
//
// Returns:
//   - a pointer to the result of applying `get` on `src`, or nil if `src` is nil
//
// Example:
//
//	name := SafePtr(user, func(u *User) string { return u.Name })
//	if name != nil {
//	    fmt.Println(*name)
//	}
//
// Use case:
//
//	Useful for safely navigating optional nested structs without panicking on nil.
func SafePtr[T any, R any](src *T, get func(*T) R) *R {
	if src == nil {
		return nil
	}

	val := get(src)
	return &val
}

// Ternary returns one of two values based on the provided boolean condition.
// Equivalent to the ternary operator (condition ? ifTrue : ifFalse) in other languages.
//
// Parameters:
//   - condition: a boolean expression
//   - ifTrue: value of type T returned if condition is true
//   - ifFalse: value of type T returned if condition is false
//
// Returns:
//   - `ifTrue` if condition is true, otherwise `ifFalse`
//
// Example:
//
//	status := Ternary(isActive, "active", "inactive")
func Ternary[T any](condition bool, ifTrue, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}

func ToPtr[T any](v T) *T {
	return &v
}

func FromPtr[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
