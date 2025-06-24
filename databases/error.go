package databases

import "errors"

var ErrNoRowFound = errors.New("no rows found in result set")
var ErrNoUpdateRow = errors.New("no rows were updated")
var ErrNoDeleteRow = errors.New("no rows were deleted")
