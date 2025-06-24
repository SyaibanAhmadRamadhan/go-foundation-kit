package generic

func TransformSlice[In any, Out any](input []In, transform func(In) Out) []Out {
	output := make([]Out, len(input))
	for i, item := range input {
		output[i] = transform(item)
	}
	return output
}

func ReverseSlice[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
