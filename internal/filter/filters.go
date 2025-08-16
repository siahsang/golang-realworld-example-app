package filter

import "github.com/siahsang/blog/internal/validator"

type Filter struct {
	Limit  int64
	Offset int64
}

type Metadata struct {
	ArticlesCount int64
}

func NewFilter(limit, offset int64) Filter {
	return Filter{
		Limit:  limit,
		Offset: offset,
	}
}

func ValidateFilters(filters Filter) *validator.Validator {
	v := validator.New()
	v.Check(filters.Limit > 0, "limit", "must be greater than 0")
	v.Check(filters.Limit <= 100, "limit", "must be a maximum of 100")
	v.Check(filters.Offset >= 0, "offset", "must be greater than or equal to 0")
	v.Check(filters.Offset <= 10_000_000, "offset", "must be a maximum of 10_000_000")

	return v
}
