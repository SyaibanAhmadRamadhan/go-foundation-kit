package filter

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
)

// BooleanFilter represents a fluent filter builder for boolean columns in SQL queries.
// Use NewBooleanFilter() to create an instance and chain methods to build filter conditions.
//
// Example:
//
//	filter := NewBooleanFilter().
//	    Column("is_active").
//	    Eq(true)
//	condition, args := filter.Build()
type BooleanFilter struct {
	column string

	// Equality operators
	eqValue  *bool
	neqValue *bool

	// Null checks
	isNull    bool
	isNotNull bool

	// Set membership
	inValues    []bool
	notInValues []bool
}

// NewBooleanFilter creates a new BooleanFilter instance.
// Chain the Column() method to set the column name, then add conditions.
//
// Returns:
//   - *BooleanFilter: A new BooleanFilter instance
//
// Example:
//
//	filter := NewBooleanFilter().
//	    Column("is_verified").
//	    Eq(true)
func NewBooleanFilter() *BooleanFilter {
	return &BooleanFilter{}
}

// Column sets the column name for this filter.
// This should be called before adding any conditions.
//
// Parameters:
//   - column: the name of the boolean column to filter
//
// Returns:
//   - *BooleanFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("is_active")
func (f *BooleanFilter) Column(column string) *BooleanFilter {
	f.column = column
	return f
}

// Eq sets an equality condition (column = value).
//
// Parameters:
//   - value: the boolean value to compare against
//
// Returns:
//   - *BooleanFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("is_active").Eq(true)
//	// SQL: is_active = true
func (f *BooleanFilter) Eq(value bool) *BooleanFilter {
	f.eqValue = &value
	return f
}

// Neq sets a not-equal condition (column <> value).
//
// Parameters:
//   - value: the boolean value to compare against
//
// Returns:
//   - *BooleanFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("is_deleted").Neq(true)
//	// SQL: is_deleted <> true (equivalent to is_deleted = false)
func (f *BooleanFilter) Neq(value bool) *BooleanFilter {
	f.neqValue = &value
	return f
}

// In sets an IN condition with multiple boolean values.
// This is useful when you want to match against multiple boolean states.
//
// Parameters:
//   - values: variadic boolean values to match against
//
// Returns:
//   - *BooleanFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("status").In(true, false)
//	// SQL: status IN (true, false)
func (f *BooleanFilter) In(values ...bool) *BooleanFilter {
	f.inValues = values
	return f
}

// NotIn sets a NOT IN condition with multiple boolean values.
//
// Parameters:
//   - values: variadic boolean values to exclude
//
// Returns:
//   - *BooleanFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("is_enabled").NotIn(false)
//	// SQL: is_enabled NOT IN (false)
func (f *BooleanFilter) NotIn(values ...bool) *BooleanFilter {
	f.notInValues = values
	return f
}

// IsNull sets an IS NULL condition.
// This checks if the boolean column value is NULL (not set).
//
// Returns:
//   - *BooleanFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("is_verified").IsNull()
//	// SQL: is_verified IS NULL
func (f *BooleanFilter) IsNull() *BooleanFilter {
	f.isNull = true
	return f
}

// IsNotNull sets an IS NOT NULL condition.
// This checks if the boolean column has a value (not NULL).
//
// Returns:
//   - *BooleanFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("is_active").IsNotNull()
//	// SQL: is_active IS NOT NULL
func (f *BooleanFilter) IsNotNull() *BooleanFilter {
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
//	filter := NewBooleanFilter().
//	    Column("is_active").
//	    Eq(true).
//	    IsNotNull()
//	condition, args := filter.Build()
//	// condition: "is_active = ? AND is_active IS NOT NULL"
//	// args: []any{true}
func (f *BooleanFilter) Build() (condition string, args []any) {
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
//	filter := NewBooleanFilter().
//	    Column("is_active").
//	    Eq(true)
//	sqlizer, _ := filter.BuildSquirrel()
//	query := squirrel.Select("*").From("users").Where(sqlizer)
func (f *BooleanFilter) BuildSquirrel() (squirrel.Sqlizer, error) {
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
