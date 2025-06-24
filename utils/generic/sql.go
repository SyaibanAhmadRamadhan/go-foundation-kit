package generic

func SqlNullTypeToPtr[T comparable](value T, isValid bool) *T {
	if !isValid {
		return nil
	}
	return &value
}
