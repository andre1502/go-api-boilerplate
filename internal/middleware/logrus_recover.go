package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// CustomRecover is a custom Echo middleware for panic recovery with Logrus logging.
func (m *Middleware) CustomRecover(logger *logrus.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					stack := debug.Stack() // Get the stack trace

					// Log the panic with Logrus
					logger.WithFields(logrus.Fields{
						"method": c.Request().Method,
						"uri":    c.Request().RequestURI,
						"ip":     c.Request().RemoteAddr,
						"error":  err.Error(), // Use err.Error() to log the string representation of the error
						"stack":  string(stack),
					}).Error("Panic recovered")

					// Return an internal server error to the client
					if c.Response().Committed {
						return
					}
					c.Error(echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error"))
				}
			}()
			return next(c)
		}
	}
}
