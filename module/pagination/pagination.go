package pagination

import (
	"context"
	"go-api-boilerplate/module"
	"math"

	"gorm.io/gorm"
)

type Pagination struct {
	TotalRecord uint64 `json:"total_record"`
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
	TotalPage   int    `json:"total_page"`
}

type paginationKey struct{}

func NewPagination() *Pagination {
	return &Pagination{
		TotalRecord: 0,
		Page:        0,
		PageSize:    module.PAGE_SIZE,
		TotalPage:   0,
	}
}

func (p *Pagination) CalculatePagination(totalRecord uint64) *Pagination {
	if (p != nil) && (p.Page > 0) && (p.PageSize > 0) && (totalRecord > 0) {
		totalPage := int(math.Ceil(float64((totalRecord)) / float64((p.PageSize))))

		p.TotalRecord = totalRecord
		p.TotalPage = totalPage

		return p
	}

	return nil
}

func (p *Pagination) Paginate(db *gorm.DB) *gorm.DB {
	if p == nil {
		return db
	}

	page := p.Page

	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * p.PageSize

	return db.Offset(offset).Limit(p.PageSize)
}

func (p *Pagination) NewContextWithPagination(ctx context.Context, page, pageSize int) context.Context {
	p.Page = page
	p.PageSize = pageSize

	return context.WithValue(ctx, paginationKey{}, p)
}

func (p *Pagination) PaginationFromContext(ctx context.Context) *Pagination {
	pCtx := ctx.Value(paginationKey{})

	if pCtx == nil {
		return nil
	}

	return pCtx.(*Pagination)
}
