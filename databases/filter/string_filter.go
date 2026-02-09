package filter

import (
	"fmt"

	"github.com/Masterminds/squirrel"
)

// StringFilter represents a fluent filter builder for string columns in SQL queries.
// Use New() to create an instance and chain methods to build filter conditions.
type StringFilter struct {
	column string

	// Equality operators
	eqValue  *string
	neqValue *string

	// Pattern matching
	likeValue     *string
	notLikeValue  *string
	iLikeValue    *string
	notILikeValue *string

	// Set membership
	inValues    []string
	notInValues []string

	// Null checks
	isNull    bool
	isNotNull bool

	// Comparison operators
	gtValue  *string
	ltValue  *string
	gteValue *string
	lteValue *string

	// Range operators
	betweenStart *string
	betweenEnd   *string
	notBetween   bool
}

// New creates a new StringFilter for the given column name.
func NewStringFilter() *StringFilter {
	return &StringFilter{}
}

// Column set an column name
func (f *StringFilter) Column(column string) *StringFilter {
	f.column = column
	return f
}

// Eq sets an equality condition (column = value).
func (f *StringFilter) Eq(value string) *StringFilter {
	f.eqValue = &value
	return f
}

// Neq sets a not-equal condition (column <> value).
func (f *StringFilter) Neq(value string) *StringFilter {
	f.neqValue = &value
	return f
}

// Like sets a LIKE pattern match condition.
func (f *StringFilter) Like(pattern string) *StringFilter {
	f.likeValue = &pattern
	return f
}

// NotLike sets a NOT LIKE pattern match condition.
func (f *StringFilter) NotLike(pattern string) *StringFilter {
	f.notLikeValue = &pattern
	return f
}

// ILike sets a case-insensitive LIKE pattern match (PostgreSQL).
func (f *StringFilter) ILike(pattern string) *StringFilter {
	f.iLikeValue = &pattern
	return f
}

// NotILike sets a case-insensitive NOT LIKE pattern match (PostgreSQL).
func (f *StringFilter) NotILike(pattern string) *StringFilter {
	f.notILikeValue = &pattern
	return f
}

// In sets an IN condition with multiple values.
func (f *StringFilter) In(values ...string) *StringFilter {
	f.inValues = values
	return f
}

// NotIn sets a NOT IN condition with multiple values.
func (f *StringFilter) NotIn(values ...string) *StringFilter {
	f.notInValues = values
	return f
}

// IsNull sets an IS NULL condition.
func (f *StringFilter) IsNull() *StringFilter {
	f.isNull = true
	return f
}

// IsNotNull sets an IS NOT NULL condition.
func (f *StringFilter) IsNotNull() *StringFilter {
	f.isNotNull = true
	return f
}

// Gt sets a greater than condition (column > value).
func (f *StringFilter) Gt(value string) *StringFilter {
	f.gtValue = &value
	return f
}

// Lt sets a less than condition (column < value).
func (f *StringFilter) Lt(value string) *StringFilter {
	f.ltValue = &value
	return f
}

// Gte sets a greater than or equal condition (column >= value).
func (f *StringFilter) Gte(value string) *StringFilter {
	f.gteValue = &value
	return f
}

// Lte sets a less than or equal condition (column <= value).
func (f *StringFilter) Lte(value string) *StringFilter {
	f.lteValue = &value
	return f
}

// Between sets a BETWEEN condition (column BETWEEN start AND end).
func (f *StringFilter) Between(start, end string) *StringFilter {
	f.betweenStart = &start
	f.betweenEnd = &end
	f.notBetween = false
	return f
}

// NotBetween sets a NOT BETWEEN condition.
func (f *StringFilter) NotBetween(start, end string) *StringFilter {
	f.betweenStart = &start
	f.betweenEnd = &end
	f.notBetween = true
	return f
}

// Build returns the SQL condition string and arguments for use with prepared statements.
// Returns empty string if no conditions are set.
func (f *StringFilter) Build() (condition string, args []any) {
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

	if f.likeValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", f.column))
		args = append(args, *f.likeValue)
	}

	if f.notLikeValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s NOT LIKE ?", f.column))
		args = append(args, *f.notLikeValue)
	}

	if f.iLikeValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s ILIKE ?", f.column))
		args = append(args, *f.iLikeValue)
	}

	if f.notILikeValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s NOT ILIKE ?", f.column))
		args = append(args, *f.notILikeValue)
	}

	if len(f.inValues) > 0 {
		placeholders := ""
		for i := range f.inValues {
			if i > 0 {
				placeholders += ", "
			}
			placeholders += "?"
			args = append(args, f.inValues[i])
		}
		conditions = append(conditions, fmt.Sprintf("%s IN (%s)", f.column, placeholders))
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

func (f *StringFilter) BuildSquirrel() (squirrel.Sqlizer, error) {
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

	// LIKE / NOT LIKE
	if f.likeValue != nil {
		where = append(where, squirrel.Like{f.column: *f.likeValue})
	}
	if f.notLikeValue != nil {
		where = append(where, squirrel.NotLike{f.column: *f.notLikeValue})
	}

	// ILIKE / NOT ILIKE (Postgres)
	if f.iLikeValue != nil {
		where = append(where, squirrel.ILike{f.column: *f.iLikeValue})
	}
	if f.notILikeValue != nil {
		where = append(where, squirrel.NotILike{f.column: *f.notILikeValue})
	}

	// IN / NOT IN
	if len(f.inValues) > 0 {
		// kalau inValues tipe []string
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
				squirrel.Gt{f.column: *f.betweenStart},
				squirrel.Lt{f.column: *f.betweenEnd},
			})
		} else {
			where = append(where, squirrel.And{
				squirrel.GtOrEq{f.column: *f.betweenStart},
				squirrel.LtOrEq{f.column: *f.betweenEnd},
			})
		}
	}

	// kalau kosong, return nil (caller bisa skip Where)
	if len(where) == 0 {
		return nil, nil
	}

	// gabung dengan AND
	return squirrel.And(where), nil
}
