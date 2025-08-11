package web

type OrderBy int

const (
	OrderAscending  OrderBy = 0
	OrderDescending OrderBy = 1
)

// ListRequest defines options related to paging and ordering to be used when listing objects
type ListRequest struct {
	Page    int     `json:"-" qs:"page" validate:"required|int|min:1|" label:"page"`
	Limit   int     `json:"-" qs:"limit" validate:"required|int|min:1|max:1000" label:"limit"`
	SortBy  string  `json:"-" qs:"sort_by" label:"sort_by"`
	OrderBy OrderBy `json:"-" qs:"order_by" validate:"int|in:0,1" label:"order_by"`
}
