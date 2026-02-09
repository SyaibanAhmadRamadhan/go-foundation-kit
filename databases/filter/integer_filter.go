package filter

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
)

// Integer is a constraint that permits any integer type.
// It includes signed integers (int, int8, int16, int32, int64) and
// unsigned integers (uint, uint8, uint16, uint32, uint64).
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// IntegerFilter represents a fluent filter builder for integer columns in SQL queries.
// It uses Go generics to support all integer types (int, int8, int16, int32, int64, uint, etc.).
// Use NewIntegerFilter() to create an instance and chain methods to build filter conditions.
//
// Example:
//
//	filter := NewIntegerFilter[int64]().
//	    Column("user_id").
//	    Eq(123)
//	condition, args := filter.Build()
type IntegerFilter[T Integer] struct {
	column string

	// Equality operators
	eqValue  *T
	neqValue *T

	// Set membership
	inValues    []T
	notInValues []T

	// Null checks
	isNull    bool
	isNotNull bool

	// Comparison operators
	gtValue  *T
	ltValue  *T
	gteValue *T
	lteValue *T

	// Range operators
	betweenStart *T
	betweenEnd   *T
	notBetween   bool
}

// NewIntegerFilter creates a new IntegerFilter instance for the specified integer type.
// The type parameter T must be one of: int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64.
//
// Type Parameter:
//   - T: the integer type for this filter (e.g., int64, int32, uint, etc.)
//
// Returns:
//   - *IntegerFilter[T]: A new IntegerFilter instance
//
// Example:
//
//	// For int64 columns
//	filter := NewIntegerFilter[int64]().
//	    Column("user_id").
//	    Eq(123)
//
//	// For int32 columns
//	filter := NewIntegerFilter[int32]().
//	    Column("age").
//	    Gte(18)
//
//	// For uint columns
//	filter := NewIntegerFilter[uint]().
//	    Column("count").
//	    Gt(0)
func NewIntegerFilter[T Integer]() *IntegerFilter[T] {
	return &IntegerFilter[T]{}
}

// Column sets the column name for this filter.
// This should be called before adding any conditions.
//
// Parameters:
//   - column: the name of the integer column to filter
//
// Returns:
//   - *IntegerFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("user_id")
func (f *IntegerFilter[T]) Column(column string) *IntegerFilter[T] {
	f.column = column
	return f
}

// Eq sets an equality condition (column = value).
//
// Parameters:
//   - value: the integer value to compare against
//
// Returns:
//   - *IntegerFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("user_id").Eq(123)
//	// SQL: user_id = ?
//	// Args: [123]
func (f *IntegerFilter[T]) Eq(value T) *IntegerFilter[T] {
	f.eqValue = &value
	return f
}

// Neq sets a not-equal condition (column <> value).
//
// Parameters:
//   - value: the integer value to compare against
//
// Returns:
//   - *IntegerFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("status").Neq(0)
//	// SQL: status <> ?
//	// Args: [0]
func (f *IntegerFilter[T]) Neq(value T) *IntegerFilter[T] {
	f.neqValue = &value
	return f
}

// In sets an IN condition with multiple integer values.
//
// Parameters:
//   - values: variadic integer values to match against
//
// Returns:
//   - *IntegerFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("role_id").In(1, 2, 3)
//	// SQL: role_id IN (?, ?, ?)
//	// Args: [1, 2, 3]
func (f *IntegerFilter[T]) In(values ...T) *IntegerFilter[T] {
	f.inValues = values
	return f
}

// NotIn sets a NOT IN condition with multiple integer values.
//
// Parameters:
//   - values: variadic integer values to exclude
//
// Returns:
//   - *IntegerFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("status_id").NotIn(0, -1)
//	// SQL: status_id NOT IN (?, ?)
//	// Args: [0, -1]
func (f *IntegerFilter[T]) NotIn(values ...T) *IntegerFilter[T] {
	f.notInValues = values
	return f
}

// IsNull sets an IS NULL condition.
// This checks if the integer column value is NULL (not set).
//
// Returns:
//   - *IntegerFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("parent_id").IsNull()
//	// SQL: parent_id IS NULL
func (f *IntegerFilter[T]) IsNull() *IntegerFilter[T] {
	f.isNull = true
	return f
}

// IsNotNull sets an IS NOT NULL condition.
// This checks if the integer column has a value (not NULL).
//
// Returns:
//   - *IntegerFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("user_id").IsNotNull()
//	// SQL: user_id IS NOT NULL
func (f *IntegerFilter[T]) IsNotNull() *IntegerFilter[T] {
	f.isNotNull = true
	return f
}

// Gt sets a greater than condition (column > value).
//
// Parameters:
//   - value: the integer value to compare against
//
// Returns:
//   - *IntegerFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("age").Gt(18)
//	// SQL: age > ?
//	// Args: [18]
func (f *IntegerFilter[T]) Gt(value T) *IntegerFilter[T] {
	f.gtValue = &value
	return f
}

// Lt sets a less than condition (column < value).
//
// Parameters:
//   - value: the integer value to compare against
//
// Returns:
//   - *IntegerFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("quantity").Lt(100)
//	// SQL: quantity < ?
//	// Args: [100]
func (f *IntegerFilter[T]) Lt(value T) *IntegerFilter[T] {
	f.ltValue = &value
	return f
}

// Gte sets a greater than or equal condition (column >= value).
//
// Parameters:
//   - value: the integer value to compare against
//
// Returns:
//   - *IntegerFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("age").Gte(18)
//	// SQL: age >= ?
//	// Args: [18]
func (f *IntegerFilter[T]) Gte(value T) *IntegerFilter[T] {
	f.gteValue = &value
	return f
}

// Lte sets a less than or equal condition (column <= value).
//
// Parameters:
//   - value: the integer value to compare against
//
// Returns:
//   - *IntegerFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("priority").Lte(5)
//	// SQL: priority <= ?
//	// Args: [5]
func (f *IntegerFilter[T]) Lte(value T) *IntegerFilter[T] {
	f.lteValue = &value
	return f
}

// Between sets a BETWEEN condition (column BETWEEN start AND end).
// The range is inclusive on both ends.
//
// Parameters:
//   - start: the starting value of the range
//   - end: the ending value of the range
//
// Returns:
//   - *IntegerFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("age").Between(18, 65)
//	// SQL: age BETWEEN ? AND ?
//	// Args: [18, 65]
func (f *IntegerFilter[T]) Between(start, end T) *IntegerFilter[T] {
	f.betweenStart = &start
	f.betweenEnd = &end
	f.notBetween = false
	return f
}

// NotBetween sets a NOT BETWEEN condition (column NOT BETWEEN start AND end).
// The range is inclusive on both ends.
//
// Parameters:
//   - start: the starting value of the range to exclude
//   - end: the ending value of the range to exclude
//
// Returns:
//   - *IntegerFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("code").NotBetween(100, 200)
//	// SQL: code NOT BETWEEN ? AND ?
//	// Args: [100, 200]
func (f *IntegerFilter[T]) NotBetween(start, end T) *IntegerFilter[T] {
	f.betweenStart = &start
	f.betweenEnd = &end
	f.notBetween = true
	return f
}

// Build returns the SQL condition string and arguments for use with prepared statements.
// Multiple conditions are combined with AND.
// Returns "1=1" if no conditions are set (always true condition).
//
// Returns:
//   - condition: the SQL WHERE condition string with ? placeholders
//   - args: slice of arguments to be used with the prepared statement
//
// Example:
//
//	filter := NewIntegerFilter[int64]().
//	    Column("user_id").
//	    Gte(100).
//	    Lte(999).
//	    IsNotNull()
//	condition, args := filter.Build()
//	// condition: "user_id >= ? AND user_id <= ? AND user_id IS NOT NULL"
//	// args: []any{100, 999}
func (f *IntegerFilter[T]) Build() (condition string, args []any) {
	var conditions []string
	args = []any{}

	if f.isNull {
		conditions = append(conditions, fmt.Sprintf("%s IS NULL", f.column))
	}

	if f.isNotNull {
		conditions = append(conditions, fmt.Sprintf("%s IS NOT NULL", f.column))
	}

	if f.eqValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s = ?", f.column))
		args = append(args, *f.eqValue)
	}

	if f.neqValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s <> ?", f.column))
		args = append(args, *f.neqValue)
	}

	if len(f.inValues) > 0 {
		var placeholders strings.Builder
		for i := range f.inValues {
			if i > 0 {
				placeholders.WriteString(", ")
			}
			placeholders.WriteString("?")
			args = append(args, f.inValues[i])
		}
		conditions = append(conditions, fmt.Sprintf("%s IN (%s)", f.column, placeholders.String()))
	}

	if len(f.notInValues) > 0 {
		placeholders := ""
		for i := range f.notInValues {
			if i > 0 {
				placeholders += ", "
			}
			placeholders += "?"
			args = append(args, f.notInValues[i])
		}
		conditions = append(conditions, fmt.Sprintf("%s NOT IN (%s)", f.column, placeholders))
	}

	if f.gtValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s > ?", f.column))
		args = append(args, *f.gtValue)
	}

	if f.ltValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s < ?", f.column))
		args = append(args, *f.ltValue)
	}

	if f.gteValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s >= ?", f.column))
		args = append(args, *f.gteValue)
	}

	if f.lteValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s <= ?", f.column))
		args = append(args, *f.lteValue)
	}

	if f.betweenStart != nil && f.betweenEnd != nil {
		if f.notBetween {
			conditions = append(conditions, fmt.Sprintf("%s NOT BETWEEN ? AND ?", f.column))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s BETWEEN ? AND ?", f.column))
		}
		args = append(args, *f.betweenStart, *f.betweenEnd)
	}

	if len(conditions) == 0 {
		return "1=1", nil
	}

	// Join multiple conditions with AND
	condition = conditions[0]
	for i := 1; i < len(conditions); i++ {
		condition += " AND " + conditions[i]
	}

	return condition, args
}

// BuildSquirrel returns a squirrel.Sqlizer for use with the squirrel query builder.
// Multiple conditions are combined with AND.
// Returns nil if no conditions are set.
//
// Returns:
//   - squirrel.Sqlizer: a squirrel condition that can be used with Where()
//   - error: always nil in current implementation
//
// Example:
//
//	filter := NewIntegerFilter[int64]().
//	    Column("user_id").
//	    Eq(123).
//	    IsNotNull()
//	sqlizer, _ := filter.BuildSquirrel()
//	query := squirrel.Select("*").From("users").Where(sqlizer)
func (f *IntegerFilter[T]) BuildSquirrel() (squirrel.Sqlizer, error) {
	where := make([]squirrel.Sqlizer, 0)

	// IS NULL / IS NOT NULL
	if f.isNull {
		where = append(where, squirrel.Expr(f.column+" IS NULL"))
	}
	if f.isNotNull {
		where = append(where, squirrel.Expr(f.column+" IS NOT NULL"))
	}

	// = / <>
	if f.eqValue != nil {
		where = append(where, squirrel.Eq{f.column: *f.eqValue})
	}
	if f.neqValue != nil {
		where = append(where, squirrel.NotEq{f.column: *f.neqValue})
	}

	// IN / NOT IN
	if len(f.inValues) > 0 {
		vals := make([]any, 0, len(f.inValues))
		for _, v := range f.inValues {
			vals = append(vals, v)
		}
		where = append(where, squirrel.Eq{f.column: vals})
	}
	if len(f.notInValues) > 0 {
		vals := make([]any, 0, len(f.notInValues))
		for _, v := range f.notInValues {
			vals = append(vals, v)
		}
		where = append(where, squirrel.NotEq{f.column: vals}) // NotEq + slice => NOT IN
	}

	// Comparison operators
	if f.gtValue != nil {
		where = append(where, squirrel.Gt{f.column: *f.gtValue})
	}
	if f.ltValue != nil {
		where = append(where, squirrel.Lt{f.column: *f.ltValue})
	}
	if f.gteValue != nil {
		where = append(where, squirrel.GtOrEq{f.column: *f.gteValue})
	}
	if f.lteValue != nil {
		where = append(where, squirrel.LtOrEq{f.column: *f.lteValue})
	}

	// BETWEEN / NOT BETWEEN
	if f.betweenStart != nil && f.betweenEnd != nil {
		if f.notBetween {
			where = append(where, squirrel.Or{
				squirrel.Lt{f.column: *f.betweenStart},
				squirrel.Gt{f.column: *f.betweenEnd},
			})
		} else {
			where = append(where, squirrel.And{
				squirrel.GtOrEq{f.column: *f.betweenStart},
				squirrel.LtOrEq{f.column: *f.betweenEnd},
			})
		}
	}

	// Return nil if no conditions (caller can skip Where)
	if len(where) == 0 {
		return nil, nil
	}

	// Combine with AND
	return squirrel.And(where), nil
}
