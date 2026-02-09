package filter

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
)

// EnumFilter represents a fluent filter builder for enum columns in SQL queries.
// It uses Go generics to support any comparable type for enum values.
// Use NewEnumFilter() to create an instance and chain methods to build filter conditions.
//
// Example with string enum:
//
//	type Status string
//	const (
//	    StatusActive   Status = "active"
//	    StatusInactive Status = "inactive"
//	)
//	filter := NewEnumFilter[Status]().
//	    Column("status").
//	    Eq(StatusActive)
//	condition, args := filter.Build()
//
// Example with integer enum:
//
//	type Role int
//	const (
//	    RoleAdmin Role = 1
//	    RoleUser  Role = 2
//	)
//	filter := NewEnumFilter[Role]().
//	    Column("role").
//	    In(RoleAdmin, RoleUser)
type EnumFilter[T comparable] struct {
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
}

// NewEnumFilter creates a new EnumFilter instance for the specified comparable type.
// The type parameter T must be a comparable type (typically string or integer-based enums).
//
// Type Parameter:
//   - T: the enum type for this filter (must be comparable)
//
// Returns:
//   - *EnumFilter[T]: A new EnumFilter instance
//
// Example:
//
//	type Status string
//	const (
//	    StatusActive   Status = "active"
//	    StatusInactive Status = "inactive"
//	)
//	filter := NewEnumFilter[Status]().
//	    Column("status").
//	    Eq(StatusActive)
func NewEnumFilter[T comparable]() *EnumFilter[T] {
	return &EnumFilter[T]{}
}

// Column sets the column name for this filter.
// This should be called before adding any conditions.
//
// Parameters:
//   - column: the name of the enum column to filter
//
// Returns:
//   - *EnumFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("status")
func (f *EnumFilter[T]) Column(column string) *EnumFilter[T] {
	f.column = column
	return f
}

// Eq sets an equality condition (column = value).
//
// Parameters:
//   - value: the enum value to compare against
//
// Returns:
//   - *EnumFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("status").Eq(StatusActive)
//	// SQL: status = ?
//	// Args: ["active"]
func (f *EnumFilter[T]) Eq(value T) *EnumFilter[T] {
	f.eqValue = &value
	return f
}

// Neq sets a not-equal condition (column <> value).
//
// Parameters:
//   - value: the enum value to compare against
//
// Returns:
//   - *EnumFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("status").Neq(StatusInactive)
//	// SQL: status <> ?
//	// Args: ["inactive"]
func (f *EnumFilter[T]) Neq(value T) *EnumFilter[T] {
	f.neqValue = &value
	return f
}

// In sets an IN condition with multiple enum values.
//
// Parameters:
//   - values: variadic enum values to match against
//
// Returns:
//   - *EnumFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("status").In(StatusActive, StatusPending)
//	// SQL: status IN (?, ?)
//	// Args: ["active", "pending"]
func (f *EnumFilter[T]) In(values ...T) *EnumFilter[T] {
	f.inValues = values
	return f
}

// NotIn sets a NOT IN condition with multiple enum values.
//
// Parameters:
//   - values: variadic enum values to exclude
//
// Returns:
//   - *EnumFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("status").NotIn(StatusDeleted, StatusArchived)
//	// SQL: status NOT IN (?, ?)
//	// Args: ["deleted", "archived"]
func (f *EnumFilter[T]) NotIn(values ...T) *EnumFilter[T] {
	f.notInValues = values
	return f
}

// IsNull sets an IS NULL condition.
// This checks if the enum column value is NULL (not set).
//
// Returns:
//   - *EnumFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("status").IsNull()
//	// SQL: status IS NULL
func (f *EnumFilter[T]) IsNull() *EnumFilter[T] {
	f.isNull = true
	return f
}

// IsNotNull sets an IS NOT NULL condition.
// This checks if the enum column has a value (not NULL).
//
// Returns:
//   - *EnumFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("status").IsNotNull()
//	// SQL: status IS NOT NULL
func (f *EnumFilter[T]) IsNotNull() *EnumFilter[T] {
	f.isNotNull = true
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
//	filter := NewEnumFilter[Status]().
//	    Column("status").
//	    In(StatusActive, StatusPending).
//	    IsNotNull()
//	condition, args := filter.Build()
//	// condition: "status IN (?, ?) AND status IS NOT NULL"
//	// args: []any{"active", "pending"}
func (f *EnumFilter[T]) Build() (condition string, args []any) {
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
//	filter := NewEnumFilter[Status]().
//	    Column("status").
//	    Eq(StatusActive)
//	sqlizer, _ := filter.BuildSquirrel()
//	query := squirrel.Select("*").From("users").Where(sqlizer)
func (f *EnumFilter[T]) BuildSquirrel() (squirrel.Sqlizer, error) {
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

	// Return nil if no conditions (caller can skip Where)
	if len(where) == 0 {
		return nil, nil
	}

	// Combine with AND
	return squirrel.And(where), nil
}
