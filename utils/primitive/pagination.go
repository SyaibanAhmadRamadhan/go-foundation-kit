package primitive

// PaginationInput holds the input parameters used for paginating results.
//
// Fields:
//   - Page: the current page number (starts from 1)
//   - PageSize: the number of items per page
type PaginationInput struct {
	Page     int64
	PageSize int64
}

// PaginationOutput provides metadata for paginated results.
//
// Fields:
//   - Page: the current page number
//   - PageSize: the number of items per page
//   - PageCount: the total number of pages available
//   - TotalData: the total number of data items found
type PaginationOutput struct {
	Page      int64
	PageSize  int64
	PageCount int64
	TotalData int64
}

// GetOffsetValue returns the offset to be used in a database query based on the page and page size.
//
// Parameters:
//   - page: the current page number (starts from 1)
//   - pageSize: number of items per page
//
// Returns:
//   - offset: number of items to skip for the query (for LIMIT/OFFSET logic)
func GetOffsetValue(page int64, pageSize int64) int64 {
	offset := int64(0)
	if page > 0 {
		offset = (page - 1) * pageSize
	}
	return offset
}

// GetPageCount calculates the total number of pages based on total data count and page size.
//
// Parameters:
//   - pageSize: number of items per page
//   - totalData: total number of items
//
// Returns:
//   - pageCount: the number of pages required to represent all data
func GetPageCount(pageSize int64, totalData int64) int64 {
	pageCount := int64(1)
	if pageSize > 0 {
		if totalData%pageSize == 0 {
			pageCount = totalData / pageSize
		} else {
			pageCount = (totalData / pageSize) + 1
		}
	}
	return pageCount
}

// CreatePaginationOutput generates a PaginationOutput object using input and total data count.
//
// Parameters:
//   - input: the PaginationInput containing page and pageSize
//   - totalData: total number of items found in the dataset
//
// Returns:
//   - PaginationOutput containing pagination metadata
func CreatePaginationOutput(input PaginationInput, totalData int64) PaginationOutput {
	pageCount := GetPageCount(input.PageSize, totalData)
	return PaginationOutput{
		Page:      input.Page,
		PageSize:  input.PageSize,
		TotalData: totalData,
		PageCount: pageCount,
	}
}
