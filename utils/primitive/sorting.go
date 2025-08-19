package primitive

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
)

type Sorting struct {
	SortFields     []string
	SortDirections []string
}

func (s Sorting) BuildSquirrel(sq squirrel.SelectBuilder) squirrel.SelectBuilder {
	for i, field := range s.SortFields {
		dir := "ASC" // default ASC
		if i < len(s.SortDirections) {
			d := s.SortDirections[i]
			if d == "desc" || d == "DESC" {
				dir = "DESC"
			}
		}
		sq = sq.OrderBy(fmt.Sprintf("%s %s", field, dir))
	}
	return sq
}

func NewSortingFromQueryParams(SortDirection, sortField string) Sorting {
	return Sorting{
		SortFields:     strings.Split(sortField, ","),
		SortDirections: strings.Split(SortDirection, ","),
	}
}
