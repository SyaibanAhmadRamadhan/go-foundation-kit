package filter

import (
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
)

// DateFilter represents a fluent filter builder for date/timestamp columns in SQL queries.
// It works with time.Time type and provides intuitive methods for date comparisons.
// Use NewDateFilter() to create an instance and chain methods to build filter conditions.
//
// Example:
//
//	filter := NewDateFilter().
//	    Column("created_at").
//	    After(time.Now().AddDate(0, -1, 0))
//	condition, args := filter.Build()
type DateFilter struct {
	column string

	// Equality operators
	eqValue  *time.Time
	neqValue *time.Time

	// Set membership
	inValues    []time.Time
	notInValues []time.Time

	// Null checks
	isNull    bool
	isNotNull bool

	// Comparison operators
	afterValue         *time.Time // column > value (Gt)
	beforeValue        *time.Time // column < value (Lt)
	afterOrEqualValue  *time.Time // column >= value (Gte)
	beforeOrEqualValue *time.Time // column <= value (Lte)

	// Range operators
	betweenStart *time.Time
	betweenEnd   *time.Time
	notBetween   bool
}

// NewDateFilter creates a new DateFilter instance for date/timestamp columns.
//
// Returns:
//   - *DateFilter: A new DateFilter instance
//
// Example:
//
//	filter := NewDateFilter().
//	    Column("created_at").
//	    After(time.Now().AddDate(0, -1, 0))
func NewDateFilter() *DateFilter {
	return &DateFilter{}
}

// Column sets the column name for this filter.
// This should be called before adding any conditions.
//
// Parameters:
//   - column: the name of the date/timestamp column to filter
//
// Returns:
//   - *DateFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("created_at")
func (f *DateFilter) Column(column string) *DateFilter {
	f.column = column
	return f
}

// Eq sets an equality condition (column = value).
// Compares the exact timestamp including time component.
// For date-only comparison, consider using Between with start and end of day.
//
// Parameters:
//   - value: the time.Time value to compare against
//
// Returns:
//   - *DateFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("created_at").Eq(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
//	// SQL: created_at = ?
func (f *DateFilter) Eq(value time.Time) *DateFilter {
	f.eqValue = &value
	return f
}

// Neq sets a not-equal condition (column <> value).
//
// Parameters:
//   - value: the time.Time value to compare against
//
// Returns:
//   - *DateFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("updated_at").Neq(time.Time{})
//	// SQL: updated_at <> ?
func (f *DateFilter) Neq(value time.Time) *DateFilter {
	f.neqValue = &value
	return f
}

// In sets an IN condition with multiple time values.
//
// Parameters:
//   - values: variadic time.Time values to match against
//
// Returns:
//   - *DateFilter: the filter instance for method chaining
//
// Example:
//
//	dates := []time.Time{
//	    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
//	    time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
//	}
//	filter.Column("event_date").In(dates...)
//	// SQL: event_date IN (?, ?)
func (f *DateFilter) In(values ...time.Time) *DateFilter {
	f.inValues = values
	return f
}

// NotIn sets a NOT IN condition with multiple time values.
//
// Parameters:
//   - values: variadic time.Time values to exclude
//
// Returns:
//   - *DateFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("scheduled_date").NotIn(holidays...)
//	// SQL: scheduled_date NOT IN (?, ?, ?)
func (f *DateFilter) NotIn(values ...time.Time) *DateFilter {
	f.notInValues = values
	return f
}

// IsNull sets an IS NULL condition.
// This checks if the date/timestamp column value is NULL (not set).
//
// Returns:
//   - *DateFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("deleted_at").IsNull()
//	// SQL: deleted_at IS NULL
func (f *DateFilter) IsNull() *DateFilter {
	f.isNull = true
	return f
}

// IsNotNull sets an IS NOT NULL condition.
// This checks if the date/timestamp column has a value (not NULL).
//
// Returns:
//   - *DateFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("created_at").IsNotNull()
//	// SQL: created_at IS NOT NULL
func (f *DateFilter) IsNotNull() *DateFilter {
	f.isNotNull = true
	return f
}

// After sets a greater than condition (column > value).
// Returns records where the date is after (later than) the specified time.
//
// Parameters:
//   - value: the time.Time value to compare against
//
// Returns:
//   - *DateFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("created_at").After(time.Now().AddDate(0, -1, 0))
//	// SQL: created_at > ?
//	// Returns records created after 1 month ago
func (f *DateFilter) After(value time.Time) *DateFilter {
	f.afterValue = &value
	return f
}

// Before sets a less than condition (column < value).
// Returns records where the date is before (earlier than) the specified time.
//
// Parameters:
//   - value: the time.Time value to compare against
//
// Returns:
//   - *DateFilter: the filter instance for method chaining
//
// Example:
//
//	filter.Column("expires_at").Before(time.Now())
//	// SQL: expires_at < ?
//	// Returns records that have expired
func (f *DateFilter) Before(value time.Time) *DateFilter {
	f.beforeValue = &value
	return f
}

// AfterOrEqual sets a greater than or equal condition (column >= value).
// Returns records where the date is on or after the specified time.
//
// Parameters:
//   - value: the time.Time value to compare against
//
// Returns:
//   - *DateFilter: the filter instance for method chaining
//
// Example:
//
//	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
//	filter.Column("created_at").AfterOrEqual(startDate)
//	// SQL: created_at >= ?
func (f *DateFilter) AfterOrEqual(value time.Time) *DateFilter {
	f.afterOrEqualValue = &value
	return f
}

// BeforeOrEqual sets a less than or equal condition (column <= value).
// Returns records where the date is on or before the specified time.
//
// Parameters:
//   - value: the time.Time value to compare against
//
// Returns:
//   - *DateFilter: the filter instance for method chaining
//
// Example:
//
//	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
//	filter.Column("created_at").BeforeOrEqual(endDate)
//	// SQL: created_at <= ?
func (f *DateFilter) BeforeOrEqual(value time.Time) *DateFilter {
	f.beforeOrEqualValue = &value
	return f
}

// Between sets a BETWEEN condition (column BETWEEN start AND end).
// The range is inclusive on both ends.
// Useful for date range queries like "records from January 2024".
//
// Parameters:
//   - start: the starting time of the range
//   - end: the ending time of the range
//
// Returns:
//   - *DateFilter: the filter instance for method chaining
//
// Example:
//
//	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
//	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
//	filter.Column("created_at").Between(startDate, endDate)
//	// SQL: created_at BETWEEN ? AND ?
//	// Returns all records from January 2024
func (f *DateFilter) Between(start, end time.Time) *DateFilter {
	f.betweenStart = &start
	f.betweenEnd = &end
	f.notBetween = false
	return f
}

// NotBetween sets a NOT BETWEEN condition (column NOT BETWEEN start AND end).
// The range is inclusive on both ends.
// Returns records outside the specified date range.
//
// Parameters:
//   - start: the starting time of the range to exclude
//   - end: the ending time of the range to exclude
//
// Returns:
//   - *DateFilter: the filter instance for method chaining
//
// Example:
//
//	// Exclude records from maintenance period
//	maintenanceStart := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
//	maintenanceEnd := time.Date(2024, 6, 7, 23, 59, 59, 0, time.UTC)
//	filter.Column("created_at").NotBetween(maintenanceStart, maintenanceEnd)
//	// SQL: created_at NOT BETWEEN ? AND ?
func (f *DateFilter) NotBetween(start, end time.Time) *DateFilter {
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
//	lastMonth := time.Now().AddDate(0, -1, 0)
//	filter := NewDateFilter().
//	    Column("created_at").
//	    After(lastMonth).
//	    IsNotNull()
//	condition, args := filter.Build()
//	// condition: "created_at > ? AND created_at IS NOT NULL"
//	// args: []any{lastMonth}
func (f *DateFilter) Build() (condition string, args []any) {
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
		var placeholders strings.Builder
		for i := range f.notInValues {
			if i > 0 {
				placeholders.WriteString(", ")
			}
			placeholders.WriteString("?")
			args = append(args, f.notInValues[i])
		}
		conditions = append(conditions, fmt.Sprintf("%s NOT IN (%s)", f.column, placeholders.String()))
	}

	if f.afterValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s > ?", f.column))
		args = append(args, *f.afterValue)
	}

	if f.beforeValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s < ?", f.column))
		args = append(args, *f.beforeValue)
	}

	if f.afterOrEqualValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s >= ?", f.column))
		args = append(args, *f.afterOrEqualValue)
	}

	if f.beforeOrEqualValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s <= ?", f.column))
		args = append(args, *f.beforeOrEqualValue)
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
//	lastWeek := time.Now().AddDate(0, 0, -7)
//	filter := NewDateFilter().
//	    Column("created_at").
//	    After(lastWeek)
//	sqlizer, _ := filter.BuildSquirrel()
//	query := squirrel.Select("*").From("orders").Where(sqlizer)
func (f *DateFilter) BuildSquirrel() (squirrel.Sqlizer, error) {
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
	if f.afterValue != nil {
		where = append(where, squirrel.Gt{f.column: *f.afterValue})
	}
	if f.beforeValue != nil {
		where = append(where, squirrel.Lt{f.column: *f.beforeValue})
	}
	if f.afterOrEqualValue != nil {
		where = append(where, squirrel.GtOrEq{f.column: *f.afterOrEqualValue})
	}
	if f.beforeOrEqualValue != nil {
		where = append(where, squirrel.LtOrEq{f.column: *f.beforeOrEqualValue})
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
