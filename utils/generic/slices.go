package generic

// TransformSlice applies a transformation function to each element in the input slice
// and returns a new slice with the transformed elements.
//
// Type Parameters:
//   - In: the type of the elements in the input slice
//   - Out: the type of the elements in the output slice
//
// Parameters:
//   - input: a slice of type In
//   - transform: a function that takes an In and returns an Out
//
// Returns:
//   - a new slice of type Out with the transformed elements
//
// Example:
//
//	numbers := []int{1, 2, 3}
//	squares := TransformSlice(numbers, func(n int) int { return n * n })
//	// squares => []int{1, 4, 9}
func TransformSlice[In any, Out any](input []In, transform func(In) Out) []Out {
	output := make([]Out, len(input))
	for i, item := range input {
		output[i] = transform(item)
	}
	return output
}

// ReverseSlice reverses the elements of a slice in-place.
//
// Type Parameters:
//   - T: the type of the elements in the slice
//
// Parameters:
//   - s: a slice of type T to be reversed
//
// Example:
//
//	names := []string{"Alice", "Bob", "Charlie"}
//	ReverseSlice(names)
//	// names => []string{"Charlie", "Bob", "Alice"}
func ReverseSlice[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// FilterSlice returns a new slice containing only the elements that satisfy the given predicate.
//
// Type Parameters:
//   - T: the type of the elements in the slice
//
// Parameters:
//   - s: a slice of type T to be filtered
//   - predicate: a function that takes an element of type T and returns a boolean indicating whether it should be included in the result
//
// Returns:
//   - a new slice of type T containing only the elements that satisfy the predicate
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	evenNumbers := FilterSlice(numbers, func(n int) bool { return n%2 == 0 })
//	// evenNumbers => []int{2, 4}
func FilterSlice[T any](s []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, item := range s {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// UniqueSlice returns a new slice containing only the unique elements of the input slice.
//
// Type Parameters:
//   - T: the type of the elements in the slice (must be comparable)
//
// Parameters:
//   - s: a slice of type T to be deduplicated
//
// Returns:
//   - a new slice of type T containing only the unique elements
//
// Example:
//
//	numbers := []int{1, 2, 2, 3, 3, 3}
//	uniqueNumbers := UniqueSlice(numbers)
//	// uniqueNumbers => []int{1, 2, 3}
func UniqueSlice[T comparable](s []T) []T {
	seen := make(map[T]bool)
	result := make([]T, 0)
	for _, item := range s {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
