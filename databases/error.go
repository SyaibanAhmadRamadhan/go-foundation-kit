package databases

import "errors"

// ErrNoRowFound is returned when a query expecting rows does not return any result.
// Common for SELECT queries using QueryRow or similar.
var ErrNoRowFound = errors.New("no rows found in result set")

// ErrNoUpdateRow is returned when an UPDATE operation affects zero rows.
// This typically indicates that the row was not found or the WHERE clause did not match anything.
var ErrNoUpdateRow = errors.New("no rows were updated")

// ErrNoDeleteRow is returned when a DELETE operation does not affect any rows.
// Usually means the specified criteria didn't match any existing record.
var ErrNoDeleteRow = errors.New("no rows were deleted")
