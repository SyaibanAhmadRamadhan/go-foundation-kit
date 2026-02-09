package filter

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
)

// Float is a constraint that permits any floating-point type.
// It includes float32 and float64.
type Float interface {
	~float32 | ~float64
}

// FloatFilter represents a fluent filter builder for floating-point columns in SQL queries.
// It uses Go generics to support both float32 and float64 types.
// Use NewFloatFilter() to create an instance and chain methods to build filter conditions.
//
// Example:
//
//	filter := NewFloatFilter[float64]().
//	    Column("price").
//	    Gte(100.50)
//	condition, args := filter.Build()
type FloatFilter[T Float] struct {
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

// NewFloatFilter creates a new FloatFilter instance for the specified floating-point type.
// The type parameter T must be one of: float32, float64.
//
// Type Parameter:
//   - T: the floating-point type for this filter (float32 or float64)
//
// Returns:
//   - *FloatFilter[T]: A new FloatFilter instance
//
// Example:
//
//	// For float64 columns
//	filter := NewFloatFilter[float64]().
//	    Column("price").
//	    Gte(99.99)
//
//	// For float32 columns
//	filter := NewFloatFilter[float32]().
//	    Column("temperature").
//	    Between(20.0, 30.0)
func NewFloatFilter[T Float]() *FloatFilter[T] {
	return &FloatFilter[T]{}
}

// Column sets the column name for this filter.
// This should be called before adding any conditions.
//
// Parameters:
//   - column: the name of the floating-point column to filter
//
// Returns:
//   - *FloatFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("price")
func (f *FloatFilter[T]) Column(column string) *FloatFilter[T] {
	f.column = column
	return f
}

// Eq sets an equality condition (column = value).
// Note: Direct equality comparison with floating-point numbers may be unreliable due to precision.
// Consider using a range comparison instead for better accuracy.
//
// Parameters:
//   - value: the floating-point value to compare against
//
// Returns:
//   - *FloatFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("price").Eq(99.99)
//	// SQL: price = ?
//	// Args: [99.99]
func (f *FloatFilter[T]) Eq(value T) *FloatFilter[T] {
	f.eqValue = &value
	return f
}

// Neq sets a not-equal condition (column <> value).
//
// Parameters:
//   - value: the floating-point value to compare against
//
// Returns:
//   - *FloatFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("discount").Neq(0.0)
//	// SQL: discount <> ?
//	// Args: [0.0]
func (f *FloatFilter[T]) Neq(value T) *FloatFilter[T] {
	f.neqValue = &value
	return f
}

// In sets an IN condition with multiple floating-point values.
//
// Parameters:
//   - values: variadic floating-point values to match against
//
// Returns:
//   - *FloatFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("rating").In(4.5, 4.8, 5.0)
//	// SQL: rating IN (?, ?, ?)
//	// Args: [4.5, 4.8, 5.0]
func (f *FloatFilter[T]) In(values ...T) *FloatFilter[T] {
	f.inValues = values
	return f
}

// NotIn sets a NOT IN condition with multiple floating-point values.
//
// Parameters:
//   - values: variadic floating-point values to exclude
//
// Returns:
//   - *FloatFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("score").NotIn(0.0, -1.0)
//	// SQL: score NOT IN (?, ?)
//	// Args: [0.0, -1.0]
func (f *FloatFilter[T]) NotIn(values ...T) *FloatFilter[T] {
	f.notInValues = values
	return f
}

// IsNull sets an IS NULL condition.
// This checks if the floating-point column value is NULL (not set).
//
// Returns:
//   - *FloatFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("commission_rate").IsNull()
//	// SQL: commission_rate IS NULL
func (f *FloatFilter[T]) IsNull() *FloatFilter[T] {
	f.isNull = true
	return f
}

// IsNotNull sets an IS NOT NULL condition.
// This checks if the floating-point column has a value (not NULL).
//
// Returns:
//   - *FloatFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("price").IsNotNull()
//	// SQL: price IS NOT NULL
func (f *FloatFilter[T]) IsNotNull() *FloatFilter[T] {
	f.isNotNull = true
	return f
}

// Gt sets a greater than condition (column > value).
//
// Parameters:
//   - value: the floating-point value to compare against
//
// Returns:
//   - *FloatFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("price").Gt(100.0)
//	// SQL: price > ?
//	// Args: [100.0]
func (f *FloatFilter[T]) Gt(value T) *FloatFilter[T] {
	f.gtValue = &value
	return f
}

// Lt sets a less than condition (column < value).
//
// Parameters:
//   - value: the floating-point value to compare against
//
// Returns:
//   - *FloatFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("discount_rate").Lt(0.5)
//	// SQL: discount_rate < ?
//	// Args: [0.5]
func (f *FloatFilter[T]) Lt(value T) *FloatFilter[T] {
	f.ltValue = &value
	return f
}

// Gte sets a greater than or equal condition (column >= value).
//
// Parameters:
//   - value: the floating-point value to compare against
//
// Returns:
//   - *FloatFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("rating").Gte(4.0)
//	// SQL: rating >= ?
//	// Args: [4.0]
func (f *FloatFilter[T]) Gte(value T) *FloatFilter[T] {
	f.gteValue = &value
	return f
}

// Lte sets a less than or equal condition (column <= value).
//
// Parameters:
//   - value: the floating-point value to compare against
//
// Returns:
//   - *FloatFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("weight").Lte(100.5)
//	// SQL: weight <= ?
//	// Args: [100.5]
func (f *FloatFilter[T]) Lte(value T) *FloatFilter[T] {
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
//   - *FloatFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("price").Between(10.99, 99.99)
//	// SQL: price BETWEEN ? AND ?
//	// Args: [10.99, 99.99]
func (f *FloatFilter[T]) Between(start, end T) *FloatFilter[T] {
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
//   - *FloatFilter[T]: the filter instance for method chaining
//
// Example:
//
//	filter.Column("score").NotBetween(0.0, 50.0)
//	// SQL: score NOT BETWEEN ? AND ?
//	// Args: [0.0, 50.0]
func (f *FloatFilter[T]) NotBetween(start, end T) *FloatFilter[T] {
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
//	filter := NewFloatFilter[float64]().
//	    Column("price").
//	    Gte(10.0).
//	    Lte(100.0).
//	    IsNotNull()
//	condition, args := filter.Build()
//	// condition: "price >= ? AND price <= ? AND price IS NOT NULL"
//	// args: []any{10.0, 100.0}
func (f *FloatFilter[T]) Build() (condition string, args []any) {
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
//	filter := NewFloatFilter[float64]().
//	    Column("price").
//	    Gte(100.0).
//	    IsNotNull()
//	sqlizer, _ := filter.BuildSquirrel()
//	query := squirrel.Select("*").From("products").Where(sqlizer)
func (f *FloatFilter[T]) BuildSquirrel() (squirrel.Sqlizer, error) {
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
