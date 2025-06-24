package generic

func SafePtr[T any, R any](src *T, get func(*T) R) *R {
	if src == nil {
		return nil
	}

	val := get(src)
	return &val
}

func Ternary[T any](condition bool, ifTrue, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}
