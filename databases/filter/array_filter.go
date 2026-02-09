package filter

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
)

// ArrayFilter represents a fluent filter builder for PostgreSQL array columns.
// It uses Go generics to support any type for array elements.
// Use NewArrayFilter() to create an instance and chain methods to build filter conditions.
//
// Supported PostgreSQL array operators:
//   - @> (contains) - Does the array contain all the specified elements?
//   - <@ (contained by) - Is the array contained by the specified array?
//   - && (overlaps) - Do the arrays have any elements in common?
//   - = (equals) - Are the arrays equal?
//   - <> (not equals) - Are the arrays not equal?
//
// Example with integer array:
//
//	filter := NewArrayFilter[int]().
//	    Column("tags").
//	    Contains(1, 2, 3)
//	condition, args := filter.Build()
//
// Example with string array:
//
//	filter := NewArrayFilter[string]().
//	    Column("categories").
//	    Overlaps("electronics", "gadgets")
//	condition, args := filter.Build()
type ArrayFilter[T any] struct {
	column string

	// Equality operators
	eqValues  []T // Array equality
	neqValues []T // Array inequality

	// PostgreSQL array operators
	containsValues    []T // @> operator
	containedByValues []T // <@ operator
	overlapsValues    []T // && operator

	// Length check
	lengthEq  *int
	lengthGt  *int
	lengthLt  *int
	lengthGte *int
	lengthLte *int

	// Null checks
	isNull    bool
	isNotNull bool

	// Empty array check
	isEmpty    bool
	isNotEmpty bool
}

// NewArrayFilter creates a new ArrayFilter instance for the specified element type.
// The type parameter T can be any type that's compatible with PostgreSQL arrays.
//
// Type Parameter:
//   - T: the element type for the array
//
// Returns:
//   - *ArrayFilter[T]: A new ArrayFilter instance
//
// Example:
//
//	// Integer array
//	filter := NewArrayFilter[int]().
//	    Column("tags").
//	    Contains(1, 2, 3)
//
//	// String array
//	filter := NewArrayFilter[string]().
//	    Column("categories").
//	    Overlaps("electronics", "gadgets")
func NewArrayFilter[T any]() *ArrayFilter[T] {
	return &ArrayFilter[T]{}
}

// Column sets the column name for this filter.
// This should be called before adding any conditions.
//
// Parameters:
//   - column: the name of the array column to filter
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags")
func (f *ArrayFilter[T]) Column(column string) *ArrayFilter[T] {
	f.column = column
	return f
}

// Eq sets an array equality condition (array = ARRAY[...]).
//
// Parameters:
//   - values: variadic elements that the array should equal
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").Eq(1, 2, 3)
//	// SQL: tags = ARRAY[?, ?, ?]
//	// Args: [1, 2, 3]
func (f *ArrayFilter[T]) Eq(values ...T) *ArrayFilter[T] {
	f.eqValues = values
	return f
}

// Neq sets an array inequality condition (array <> ARRAY[...]).
//
// Parameters:
//   - values: variadic elements that the array should not equal
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").Neq(1, 2, 3)
//	// SQL: tags <> ARRAY[?, ?, ?]
//	// Args: [1, 2, 3]
func (f *ArrayFilter[T]) Neq(values ...T) *ArrayFilter[T] {
	f.neqValues = values
	return f
}

// Contains sets a @> (contains) condition.
// Checks if the array contains all the specified elements.
//
// Parameters:
//   - values: variadic elements that should be contained in the array
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").Contains(1, 2)
//	// SQL: tags @> ARRAY[?, ?]
//	// Args: [1, 2]
func (f *ArrayFilter[T]) Contains(values ...T) *ArrayFilter[T] {
	f.containsValues = values
	return f
}

// ContainedBy sets a <@ (contained by) condition.
// Checks if the array is contained by the specified elements.
//
// Parameters:
//   - values: variadic elements that should contain the array
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").ContainedBy(1, 2, 3, 4, 5)
//	// SQL: tags <@ ARRAY[?, ?, ?, ?, ?]
//	// Args: [1, 2, 3, 4, 5]
func (f *ArrayFilter[T]) ContainedBy(values ...T) *ArrayFilter[T] {
	f.containedByValues = values
	return f
}

// Overlaps sets a && (overlaps) condition.
// Checks if the array has any elements in common with the specified elements.
//
// Parameters:
//   - values: variadic elements to check for overlap
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").Overlaps(1, 2, 3)
//	// SQL: tags && ARRAY[?, ?, ?]
//	// Args: [1, 2, 3]
func (f *ArrayFilter[T]) Overlaps(values ...T) *ArrayFilter[T] {
	f.overlapsValues = values
	return f
}

// LengthEq sets a condition to check if the array length equals the specified value.
//
// Parameters:
//   - length: the expected array length
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").LengthEq(3)
//	// SQL: array_length(tags, 1) = ?
//	// Args: [3]
func (f *ArrayFilter[T]) LengthEq(length int) *ArrayFilter[T] {
	f.lengthEq = &length
	return f
}

// LengthGt sets a condition to check if the array length is greater than the specified value.
//
// Parameters:
//   - length: the minimum array length (exclusive)
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").LengthGt(2)
//	// SQL: array_length(tags, 1) > ?
//	// Args: [2]
func (f *ArrayFilter[T]) LengthGt(length int) *ArrayFilter[T] {
	f.lengthGt = &length
	return f
}

// LengthLt sets a condition to check if the array length is less than the specified value.
//
// Parameters:
//   - length: the maximum array length (exclusive)
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").LengthLt(10)
//	// SQL: array_length(tags, 1) < ?
//	// Args: [10]
func (f *ArrayFilter[T]) LengthLt(length int) *ArrayFilter[T] {
	f.lengthLt = &length
	return f
}

// LengthGte sets a condition to check if the array length is greater than or equal to the specified value.
//
// Parameters:
//   - length: the minimum array length (inclusive)
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").LengthGte(1)
//	// SQL: array_length(tags, 1) >= ?
//	// Args: [1]
func (f *ArrayFilter[T]) LengthGte(length int) *ArrayFilter[T] {
	f.lengthGte = &length
	return f
}

// LengthLte sets a condition to check if the array length is less than or equal to the specified value.
//
// Parameters:
//   - length: the maximum array length (inclusive)
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").LengthLte(5)
//	// SQL: array_length(tags, 1) <= ?
//	// Args: [5]
func (f *ArrayFilter[T]) LengthLte(length int) *ArrayFilter[T] {
	f.lengthLte = &length
	return f
}

// IsEmpty sets a condition to check if the array is empty.
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").IsEmpty()
//	// SQL: (tags = ARRAY[]::integer[] OR array_length(tags, 1) IS NULL)
func (f *ArrayFilter[T]) IsEmpty() *ArrayFilter[T] {
	f.isEmpty = true
	return f
}

// IsNotEmpty sets a condition to check if the array is not empty.
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").IsNotEmpty()
//	// SQL: (tags <> ARRAY[]::integer[] AND array_length(tags, 1) IS NOT NULL)
func (f *ArrayFilter[T]) IsNotEmpty() *ArrayFilter[T] {
	f.isNotEmpty = true
	return f
}

// IsNull sets an IS NULL condition.
// This checks if the array column value is NULL (not set).
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").IsNull()
//	// SQL: tags IS NULL
func (f *ArrayFilter[T]) IsNull() *ArrayFilter[T] {
	f.isNull = true
	return f
}

// IsNotNull sets an IS NOT NULL condition.
// This checks if the array column has a value (not NULL).
//
// Returns:
//   - *ArrayFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("tags").IsNotNull()
//	// SQL: tags IS NOT NULL
func (f *ArrayFilter[T]) IsNotNull() *ArrayFilter[T] {
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
//	filter := NewArrayFilter[int]().
//	    Column("tags").
//	    Contains(1, 2).
//	    IsNotNull()
//	condition, args := filter.Build()
//	// condition: "tags @> ARRAY[?, ?] AND tags IS NOT NULL"
//	// args: []any{1, 2}
func (f *ArrayFilter[T]) Build() (condition string, args []any) {
	var conditions []string
	args = []any{}

	if f.isNull {
		conditions = append(conditions, fmt.Sprintf("%s IS NULL", f.column))
	}

	if f.isNotNull {
		conditions = append(conditions, fmt.Sprintf("%s IS NOT NULL", f.column))
	}

	if len(f.eqValues) > 0 {
		var placeholders strings.Builder
		placeholders.WriteString("ARRAY[")
		for i := range f.eqValues {
			if i > 0 {
				placeholders.WriteString(", ")
			}
			placeholders.WriteString("?")
			args = append(args, f.eqValues[i])
		}
		placeholders.WriteString("]")
		conditions = append(conditions, fmt.Sprintf("%s = %s", f.column, placeholders.String()))
	}

	if len(f.neqValues) > 0 {
		var placeholders strings.Builder
		placeholders.WriteString("ARRAY[")
		for i := range f.neqValues {
			if i > 0 {
				placeholders.WriteString(", ")
			}
			placeholders.WriteString("?")
			args = append(args, f.neqValues[i])
		}
		placeholders.WriteString("]")
		conditions = append(conditions, fmt.Sprintf("%s <> %s", f.column, placeholders.String()))
	}

	if len(f.containsValues) > 0 {
		var placeholders strings.Builder
		placeholders.WriteString("ARRAY[")
		for i := range f.containsValues {
			if i > 0 {
				placeholders.WriteString(", ")
			}
			placeholders.WriteString("?")
			args = append(args, f.containsValues[i])
		}
		placeholders.WriteString("]")
		conditions = append(conditions, fmt.Sprintf("%s @> %s", f.column, placeholders.String()))
	}

	if len(f.containedByValues) > 0 {
		var placeholders strings.Builder
		placeholders.WriteString("ARRAY[")
		for i := range f.containedByValues {
			if i > 0 {
				placeholders.WriteString(", ")
			}
			placeholders.WriteString("?")
			args = append(args, f.containedByValues[i])
		}
		placeholders.WriteString("]")
		conditions = append(conditions, fmt.Sprintf("%s <@ %s", f.column, placeholders.String()))
	}

	if len(f.overlapsValues) > 0 {
		var placeholders strings.Builder
		placeholders.WriteString("ARRAY[")
		for i := range f.overlapsValues {
			if i > 0 {
				placeholders.WriteString(", ")
			}
			placeholders.WriteString("?")
			args = append(args, f.overlapsValues[i])
		}
		placeholders.WriteString("]")
		conditions = append(conditions, fmt.Sprintf("%s && %s", f.column, placeholders.String()))
	}

	if f.lengthEq != nil {
		conditions = append(conditions, fmt.Sprintf("array_length(%s, 1) = ?", f.column))
		args = append(args, *f.lengthEq)
	}

	if f.lengthGt != nil {
		conditions = append(conditions, fmt.Sprintf("array_length(%s, 1) > ?", f.column))
		args = append(args, *f.lengthGt)
	}

	if f.lengthLt != nil {
		conditions = append(conditions, fmt.Sprintf("array_length(%s, 1) < ?", f.column))
		args = append(args, *f.lengthLt)
	}

	if f.lengthGte != nil {
		conditions = append(conditions, fmt.Sprintf("array_length(%s, 1) >= ?", f.column))
		args = append(args, *f.lengthGte)
	}

	if f.lengthLte != nil {
		conditions = append(conditions, fmt.Sprintf("array_length(%s, 1) <= ?", f.column))
		args = append(args, *f.lengthLte)
	}

	if f.isEmpty {
		conditions = append(conditions, fmt.Sprintf("(array_length(%s, 1) IS NULL OR array_length(%s, 1) = 0)", f.column, f.column))
	}

	if f.isNotEmpty {
		conditions = append(conditions, fmt.Sprintf("(array_length(%s, 1) IS NOT NULL AND array_length(%s, 1) > 0)", f.column, f.column))
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
//	filter := NewArrayFilter[int]().
//	    Column("tags").
//	    Contains(1, 2, 3)
//	sqlizer, _ := filter.BuildSquirrel()
//	query := squirrel.Select("*").From("items").Where(sqlizer)
func (f *ArrayFilter[T]) BuildSquirrel() (squirrel.Sqlizer, error) {
	where := make([]squirrel.Sqlizer, 0)

	// IS NULL / IS NOT NULL
	if f.isNull {
		where = append(where, squirrel.Expr(f.column+" IS NULL"))
	}
	if f.isNotNull {
		where = append(where, squirrel.Expr(f.column+" IS NOT NULL"))
	}

	// Array equality
	if len(f.eqValues) > 0 {
		vals := make([]any, len(f.eqValues))
		for i, v := range f.eqValues {
			vals[i] = v
		}
		placeholders := make([]string, len(vals))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		arrayStr := fmt.Sprintf("ARRAY[%s]", strings.Join(placeholders, ", "))
		where = append(where, squirrel.Expr(f.column+" = "+arrayStr, vals...))
	}

	// Array inequality
	if len(f.neqValues) > 0 {
		vals := make([]any, len(f.neqValues))
		for i, v := range f.neqValues {
			vals[i] = v
		}
		placeholders := make([]string, len(vals))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		arrayStr := fmt.Sprintf("ARRAY[%s]", strings.Join(placeholders, ", "))
		where = append(where, squirrel.Expr(f.column+" <> "+arrayStr, vals...))
	}

	// @> (contains)
	if len(f.containsValues) > 0 {
		vals := make([]any, len(f.containsValues))
		for i, v := range f.containsValues {
			vals[i] = v
		}
		placeholders := make([]string, len(vals))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		arrayStr := fmt.Sprintf("ARRAY[%s]", strings.Join(placeholders, ", "))
		where = append(where, squirrel.Expr(f.column+" @> "+arrayStr, vals...))
	}

	// <@ (contained by)
	if len(f.containedByValues) > 0 {
		vals := make([]any, len(f.containedByValues))
		for i, v := range f.containedByValues {
			vals[i] = v
		}
		placeholders := make([]string, len(vals))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		arrayStr := fmt.Sprintf("ARRAY[%s]", strings.Join(placeholders, ", "))
		where = append(where, squirrel.Expr(f.column+" <@ "+arrayStr, vals...))
	}

	// && (overlaps)
	if len(f.overlapsValues) > 0 {
		vals := make([]any, len(f.overlapsValues))
		for i, v := range f.overlapsValues {
			vals[i] = v
		}
		placeholders := make([]string, len(vals))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		arrayStr := fmt.Sprintf("ARRAY[%s]", strings.Join(placeholders, ", "))
		where = append(where, squirrel.Expr(f.column+" && "+arrayStr, vals...))
	}

	// Length conditions
	if f.lengthEq != nil {
		where = append(where, squirrel.Expr(fmt.Sprintf("array_length(%s, 1) = ?", f.column), *f.lengthEq))
	}
	if f.lengthGt != nil {
		where = append(where, squirrel.Expr(fmt.Sprintf("array_length(%s, 1) > ?", f.column), *f.lengthGt))
	}
	if f.lengthLt != nil {
		where = append(where, squirrel.Expr(fmt.Sprintf("array_length(%s, 1) < ?", f.column), *f.lengthLt))
	}
	if f.lengthGte != nil {
		where = append(where, squirrel.Expr(fmt.Sprintf("array_length(%s, 1) >= ?", f.column), *f.lengthGte))
	}
	if f.lengthLte != nil {
		where = append(where, squirrel.Expr(fmt.Sprintf("array_length(%s, 1) <= ?", f.column), *f.lengthLte))
	}

	// Empty/not empty checks
	if f.isEmpty {
		where = append(where, squirrel.Expr(fmt.Sprintf("(array_length(%s, 1) IS NULL OR array_length(%s, 1) = 0)", f.column, f.column)))
	}
	if f.isNotEmpty {
		where = append(where, squirrel.Expr(fmt.Sprintf("(array_length(%s, 1) IS NOT NULL AND array_length(%s, 1) > 0)", f.column, f.column)))
	}

	// Return nil if no conditions (caller can skip Where)
	if len(where) == 0 {
		return nil, nil
	}

	// Combine with AND
	return squirrel.And(where), nil
}
