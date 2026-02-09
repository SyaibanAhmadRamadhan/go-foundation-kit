package filter

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// UUIDFilter represents a fluent filter builder for UUID columns in SQL queries.
// It uses the github.com/google/uuid package for UUID type handling.
// Use NewUUIDFilter() to create an instance and chain methods to build filter conditions.
//
// Example:
//
//	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
//	filter := NewUUIDFilter().
//	    Column("user_id").
//	    Eq(id)
//	condition, args := filter.Build()
type UUIDFilter struct {
	column string

	// Equality operators
	eqValue  *uuid.UUID
	neqValue *uuid.UUID

	// Set membership
	inValues    []uuid.UUID
	notInValues []uuid.UUID

	// Null checks
	isNull    bool
	isNotNull bool
}

// NewUUIDFilter creates a new UUIDFilter instance.
// Chain the Column() method to set the column name, then add conditions.
//
// Returns:
//   - *UUIDFilter: A new UUIDFilter instance
//
// Example:
//
//	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
//	filter := NewUUIDFilter().
//	    Column("user_id").
//	    Eq(id)
func NewUUIDFilter() *UUIDFilter {
	return &UUIDFilter{}
}

// Column sets the column name for this filter.
// This should be called before adding any conditions.
//
// Parameters:
//   - column: the name of the UUID column to filter
//
// Returns:
//   - *UUIDFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("user_id")
func (f *UUIDFilter) Column(column string) *UUIDFilter {
	f.column = column
	return f
}

// Eq sets an equality condition (column = value).
//
// Parameters:
//   - value: the UUID value to compare against
//
// Returns:
//   - *UUIDFilter: the filter instance for method chaining
//
// Example:
//
//	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
//	filter.Column("user_id").Eq(id)
//	// SQL: user_id = ?
//	// Args: [550e8400-e29b-41d4-a716-446655440000]
func (f *UUIDFilter) Eq(value uuid.UUID) *UUIDFilter {
	f.eqValue = &value
	return f
}

// Neq sets a not-equal condition (column <> value).
//
// Parameters:
//   - value: the UUID value to compare against
//
// Returns:
//   - *UUIDFilter: the filter instance for method chaining
//
// Example:
//
//	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
//	filter.Column("user_id").Neq(id)
//	// SQL: user_id <> ?
//	// Args: [550e8400-e29b-41d4-a716-446655440000]
func (f *UUIDFilter) Neq(value uuid.UUID) *UUIDFilter {
	f.neqValue = &value
	return f
}

// In sets an IN condition with multiple UUID values.
//
// Parameters:
//   - values: variadic UUID values to match against
//
// Returns:
//   - *UUIDFilter: the filter instance for method chaining
//
// Example:
//
//	id1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
//	id2 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
//	filter.Column("user_id").In(id1, id2)
//	// SQL: user_id IN (?, ?)
//	// Args: [550e8400-e29b-41d4-a716-446655440000, 6ba7b810-9dad-11d1-80b4-00c04fd430c8]
func (f *UUIDFilter) In(values ...uuid.UUID) *UUIDFilter {
	f.inValues = values
	return f
}

// NotIn sets a NOT IN condition with multiple UUID values.
//
// Parameters:
//   - values: variadic UUID values to exclude
//
// Returns:
//   - *UUIDFilter: the filter instance for method chaining
//
// Example:
//
//	id1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
//	id2 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
//	filter.Column("user_id").NotIn(id1, id2)
//	// SQL: user_id NOT IN (?, ?)
//	// Args: [550e8400-e29b-41d4-a716-446655440000, 6ba7b810-9dad-11d1-80b4-00c04fd430c8]
func (f *UUIDFilter) NotIn(values ...uuid.UUID) *UUIDFilter {
	f.notInValues = values
	return f
}

// IsNull sets an IS NULL condition.
// This checks if the UUID column value is NULL (not set).
//
// Returns:
//   - *UUIDFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("parent_id").IsNull()
//	// SQL: parent_id IS NULL
func (f *UUIDFilter) IsNull() *UUIDFilter {
	f.isNull = true
	return f
}

// IsNotNull sets an IS NOT NULL condition.
// This checks if the UUID column has a value (not NULL).
//
// Returns:
//   - *UUIDFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("user_id").IsNotNull()
//	// SQL: user_id IS NOT NULL
func (f *UUIDFilter) IsNotNull() *UUIDFilter {
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
//	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
//	filter := NewUUIDFilter().
//	    Column("user_id").
//	    Eq(id).
//	    IsNotNull()
//	condition, args := filter.Build()
//	// condition: "user_id = ? AND user_id IS NOT NULL"
//	// args: []any{550e8400-e29b-41d4-a716-446655440000}
func (f *UUIDFilter) Build() (condition string, args []any) {
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
//	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
//	filter := NewUUIDFilter().
//	    Column("user_id").
//	    Eq(id)
//	sqlizer, _ := filter.BuildSquirrel()
//	query := squirrel.Select("*").From("users").Where(sqlizer)
func (f *UUIDFilter) BuildSquirrel() (squirrel.Sqlizer, error) {
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
