package middleware

import (
	"go-api-boilerplate/internal/constant"
	"go-api-boilerplate/module"

	"github.com/labstack/echo/v4"
)

func (m *Middleware) RequestIPAddress(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ipAddress := module.GetClientIP(c.Request())

		c.Set(constant.IP_ADDRESS, ipAddress)

		return next(c)
	}
}
