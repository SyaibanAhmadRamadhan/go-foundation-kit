package databases

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrNoRowFound is returned when a query expecting rows does not return any result.
// Common for SELECT queries using QueryRow or similar.
var ErrNoRowFound = errors.New("no rows found in result set")

// ErrNoUpdateRow is returned when an UPDATE operation affects zero rows.
// This typically indicates that the row was not found or the WHERE clause did not match anything.
var ErrNoUpdateRow = errors.New("no rows were updated")

// ErrNoDeleteRow is returned when a DELETE operation does not affect any rows.
// Usually means the specified criteria didn't match any existing record.
var ErrNoDeleteRow = errors.New("no rows were deleted")

// ErrDuplicateEntry is returned when an operation violates a uniqueness constraint,
// such as attempting to insert a duplicate value into a column that requires unique values.
var ErrDuplicateEntry = errors.New("duplicate entry")

// MapDuplicateEntryError inspects the provided error and maps it to ErrDuplicateEntry
// if it corresponds to a duplicate entry error from either PostgreSQL or MySQL.
// If the error does not indicate a duplicate entry, it is returned unchanged.
func MapDuplicateEntryError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return ErrDuplicateEntry
		}
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrDuplicateEntry
	}
	return err
}
