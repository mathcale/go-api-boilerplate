package database

type PaginatedResult[T any] struct {
	TotalRecords uint64 `json:"total_records"`
	TotalPages   uint64 `json:"total_pages"`
	PageSize     uint64 `json:"page_size"`
	CurrentPage  uint64 `json:"current_page"`
	PrevPage     uint64 `json:"prev_page"`
	NextPage     uint64 `json:"next_page"`
	Data         []T    `json:"data"`
}
