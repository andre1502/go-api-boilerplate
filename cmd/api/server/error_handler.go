package server

import (
	"errors"
	"go-api-boilerplate/internal/exception"
	"go-api-boilerplate/internal/status_code"
	"net/http"
	"reflect"
	"runtime"

	"github.com/labstack/echo/v4"
)

func (s *Server) CustomErrorHandler(err error, c echo.Context) {
	var message any

	httpStatus := http.StatusInternalServerError
	code := status_code.SERVER_ERROR_CODE
	message = status_code.SERVER_ERROR_MESSAGE

	if he, ok := err.(*exception.Exception); ok {
		httpStatus = http.StatusOK
		code = he.Code
		message = he.Message
	} else if he, ok := err.(*echo.HTTPError); ok {
		httpStatus = he.Code
		message = he.Message

		if he.Internal != nil {
			if herr, ok := he.Internal.(*echo.HTTPError); ok {
				httpStatus = herr.Code
				message = herr.Message
			}
		}
	} else if he, ok := err.(*runtime.TypeAssertionError); ok {
		code = status_code.TYPE_ASSERTION_ERROR_CODE
		message = status_code.TYPE_ASSERTION_ERROR_MESSAGE
		err = &CustomError{
			Message: he.Error(),
			Err:     he,
		}
	} else {
		valueErr := reflect.ValueOf(err)

		err = &CustomError{
			Message: err.Error(),
			Err:     s.getErrorFromPointer(valueErr),
		}
	}

	if !c.Response().Committed {
		if httpStatus == http.StatusNotFound {
			s.response.NotFound(c)
		} else {
			s.response.ErrorHandler(c, httpStatus, exception.Ex.Errors(code, message.(string), err))
		}
	}
}

func (s *Server) getErrorFromPointer(val reflect.Value) error {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()

		return s.getErrorFromPointer(val)
	}

	return errors.New(val.String())

}
