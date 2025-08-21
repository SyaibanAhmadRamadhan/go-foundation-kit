package databases

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/reflection"
)

// NewJSON creates a new JSON[T] wrapper from a given value.
// The Valid flag will be set to true if the value is not nil.
func NewJSON[T any](val T) JSON[T] {
	return JSON[T]{
		V:     val,
		Valid: !reflection.IsNil(val),
	}
}

// JSON is a generic type that wraps any Go type T and allows it
// to be stored/retrieved as a JSON column in SQL databases.
// It implements sql.Scanner and driver.Valuer.
type JSON[T any] struct {
	V     T    // V the underlying value
	Valid bool // Valid true if the value is valid (not NULL)
}

// Scan implements the sql.Scanner interface.
// It decodes the value from a SQL column into the JSON[T] wrapper.
func (j *JSON[T]) Scan(value any) error {
	if j == nil {
		return fmt.Errorf("JSON.Scan: receiver is nil")
	}
	if value == nil {
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = append([]byte(nil), v...)
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("JSON.Scan: unsupported source type %T", value)
	}

	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		var zero T
		j.V = zero
		j.Valid = false
		return nil
	}

	if err := json.Unmarshal(b, &j.V); err != nil {
		return fmt.Errorf("JSON.Scan: invalid JSON: %w", err)
	}
	j.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
// It encodes the underlying value into JSON for storing in a SQL column.
func (j JSON[T]) Value() (driver.Value, error) {
	if !j.Valid {
		return nil, nil
	}

	b, err := json.Marshal(j.V)
	if err != nil {
		return nil, fmt.Errorf("JSON.Value: marshal failed: %w", err)
	}

	return string(b), nil
}

// MarshalFrom sets the underlying value and updates the Valid flag.
func (j *JSON[T]) MarshalFrom(v T) {
	j.V = v
	j.Valid = !reflection.IsNil(v)
}

// Into returns the underlying value directly.
func (j JSON[T]) Into() T {
	return j.V
}

// IntoPtr returns a pointer to the underlying value.
// It returns nil if the JSON is invalid/null.
func (j JSON[T]) IntoPtr() *T {
	if !j.Valid {
		return nil
	}

	if reflection.IsZero(j.V) {
		return &j.V
	}
	out := j.V
	return &out
}

// IsZero reports whether the underlying value is the zero value for its type.
func (j JSON[T]) IsZero() bool {
	return reflection.IsZero(j.V)
}
