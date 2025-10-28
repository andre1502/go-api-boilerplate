package middleware

import (
	"go-api-boilerplate/internal/constant"

	"github.com/labstack/echo/v4"
)

func (m *Middleware) CheckJwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		jwtClaims, tkn, err := m.GetJwtClaims(c, constant.USER_JWT)
		if (err != nil) || (jwtClaims == nil) {
			return m.Response.Unauthorized(c)
		}

		c.Set(constant.USER_TOKEN, tkn)
		c.Set(constant.USER_ID, jwtClaims.UID)

		if jwtClaims.GUID != nil {
			c.Set(constant.USER_GUID, *jwtClaims.GUID)
		}

		return next(c)
	}
}
