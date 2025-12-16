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

// ToPtr returns a pointer to the given value.
// If the value is a zero value, it still returns a pointer to that zero value.
//
// Parameters:
//   - v: a value of type T
//
// Returns:
//   - a pointer to the value `v`
//
// Example:
//
//	num := 42
//	numPtr := ToPtr(num)
//	// numPtr => *int pointing to 42
func ToPtr[T any](v T) *T {
	return &v
}

// FromPtr dereferences a pointer to obtain the underlying value.
// If the pointer is nil, it returns the zero value of type T.
//
// Parameters:
//   - p: a pointer to a value of type T
//
// Returns:
//   - the value pointed to by `p`, or the zero value of type T if `p` is nil
//
// Example:
//
//	var strPtr *string
//	str := FromPtr(strPtr)
//	// str => "" (zero value of string)
func FromPtr[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// CastPtr casts a pointer of one type to a pointer of another type using a provided casting function.
// If the input pointer is nil, it returns nil.
//
// Parameters:
//   - v: a pointer to a value of type From
//   - cast: a function that takes a From and returns a To
//
// Returns:
//   - a pointer to the casted value of type To, or nil if `v` is nil
//
// Example:
//
//	var intPtr *int = ToPtr(42)
//	floatPtr := CastPtr(intPtr, func(i int) float64 { return float64(i) })
//	// floatPtr => *float64 pointing to 42.0
func CastPtr[From any, To any](v *From, cast func(From) To) *To {
	if v == nil {
		return nil
	}

	out := cast(*v)
	return &out
}

// CastConstStringPtr casts a pointer of one string-like type to another string-like type.
// If the input pointer is nil, it returns nil.
//
// Type Parameters:
//   - To: the target string-like type
//   - From: the source string-like type
//
// Parameters:
//   - v: a pointer to a value of type From
//
// Returns:
//   - a pointer to the casted value of type To, or nil if `v` is nil
//
// Example:
//
//	var myStringPtr *MyStringType = ToPtr(MyStringType("hello"))
//	standardStringPtr := CastConstStringPtr[string, MyStringType](myStringPtr)
//	// standardStringPtr => *string pointing to "hello"
func CastConstStringPtr[To, From ~string](v *From) *To {
	if v == nil {
		return nil
	}

	out := To(*v)
	return &out
}

// TransformPointer applies a transformation function to a pointer,
// returning a new pointer to the result if the source is not nil.
//
// Type Parameters:
//   - From: the type of the input pointer
//   - To: the type of the output pointer
//
// Parameters:
//   - in: a pointer to a value of type From
//   - transform: a function that takes a *From and returns a *To
//
// Returns:
//   - a pointer to the result of applying `transform` on `in`, or nil if `in` is nil
//
// Example:
//
//	var agePtr *int = ToPtr(30)
//	ageStrPtr := TransformPointer(agePtr, func(a *int) *string {
//	    str := fmt.Sprintf("%d years", *a)
//	    return &str
//	})
//	// ageStrPtr => *string pointing to "30 years"
func TransformPointer[From any, To any](in *From, transform func(*From) *To) *To {
	if in == nil {
		return nil
	}
	return transform(in)
}
