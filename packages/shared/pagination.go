package shared

import "math"

const DefaultPageSize int32 = 10

type PageRequest struct {
	page     int32 // 1-based index (Page=1 — первая страница)
	pageSize int32
}

type PageResponse struct {
	Page       int32
	PageSize   int32
	TotalPages int32
	TotalCount int32
}

func NewPageRequest(page, pageSize int32) *PageRequest {
	return &PageRequest{
		page:     page,
		pageSize: pageSize,
	}
}

func (p *PageRequest) Page() int32 {
	if p.page <= 0 {
		return 1
	}
	return p.page
}

func (p *PageRequest) PageSize() int32 {
	if p.pageSize <= 0 {
		return DefaultPageSize
	}
	return p.pageSize
}

func (p *PageRequest) Offset() int32 {
	page := p.page
	if page <= 1 {
		page = 1
	}
	pageSize := p.PageSize()
	return (page - 1) * pageSize
}

func (p *PageRequest) TotalPages(totalCount int32) int32 {
	pageSize := p.PageSize()
	if pageSize == 0 {
		return 0 // или 1, но чаще 0
	}
	return int32(math.Ceil(float64(totalCount) / float64(pageSize)))
}
