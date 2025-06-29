package generic

// SqlNullTypeToPtr converts a SQL nullable type to a Go pointer.
// If the SQL value is invalid (i.e., NULL), it returns nil; otherwise, it returns a pointer to the value.
//
// Type Parameters:
//   - T: any comparable type (e.g., string, int, time.Time)
//
// Parameters:
//   - value: the value to convert
//   - isValid: a boolean indicating whether the value is valid (i.e., not NULL)
//
// Returns:
//   - a pointer to `value` if `isValid` is true; otherwise, nil
//
// Example:
//
//	var sqlString sql.NullString
//	name := SqlNullTypeToPtr(sqlString.String, sqlString.Valid)
//	if name != nil {
//	    fmt.Println(*name)
//	}
func SqlNullTypeToPtr[T comparable](value T, isValid bool) *T {
	if !isValid {
		return nil
	}
	return &value
}
