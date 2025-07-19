package duel

// FilterType represents the type of filter to apply when getting duel list.
type FilterType string

const (
	FilterTypeAll      FilterType = "all"
	FilterType1v1      FilterType = "1v1"
	FilterTypeDailyTop FilterType = "daily_top"
	FilterTypeMyDuels  FilterType = "my_duels"
)

// Filter represents the filter parameters for getting duel list.
type Filter struct {
	FilterType FilterType
}

// NewFilter creates a new filter with the specified type.
func NewFilter(filterType FilterType) *Filter {
	return &Filter{
		FilterType: filterType,
	}
}
