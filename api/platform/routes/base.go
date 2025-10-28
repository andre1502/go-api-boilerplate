package routes

import (
	"go-api-boilerplate/api/platform/middleware"

	"github.com/labstack/echo/v4"
)

type Route struct {
	middleware *middleware.Middleware
}

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
	platformGroup := api.Group("/api")

	r.SetupV1Routes(platformGroup)
}

func (r Route) SetupV1Routes(api *echo.Group) {
	// v1Group := api.Group("/v1")
}
