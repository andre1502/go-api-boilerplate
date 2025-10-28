package routes

import (
	"go-api-boilerplate/api/backoffice/middleware"

	"github.com/labstack/echo/v4"
)

type Route struct {
	middleware *middleware.Middleware
}

// NewRoute 創建一個新的 Route
func NewRoute(
	mdl *middleware.Middleware,
) *Route {
	return &Route{
		middleware: mdl,
	}
}

func (r Route) GetHealth() bool {
	return true
}

func (r Route) SetupRoutes(api *echo.Group) {
}
