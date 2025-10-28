package middleware

import (
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/pagination"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (m *Middleware) Paginate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var page int
		var pageSize int
		var err error

		pageStr := module.GetParamDefault(c.QueryParam("page"), "0")
		pageSizeStr := module.GetParamDefault(c.QueryParam("page_size"), strconv.Itoa(module.PAGE_SIZE))

		if page, err = strconv.Atoi(pageStr); err != nil || page < 0 {
			if module.IsEmptyString(pageSizeStr) {
				page = 0
			} else {
				page = 1
			}
		}

		pageSize = module.StrToIntDefault(pageSizeStr, module.PAGE_SIZE)

		if pageSize < 0 {
			pageSize = module.PAGE_SIZE
		}

		paginationObj := pagination.NewPagination()
		ctx := paginationObj.NewContextWithPagination(c.Request().Context(), page, pageSize)
		req := c.Request().WithContext(ctx)
		c.SetRequest(req)

		return next(c)
	}
}
