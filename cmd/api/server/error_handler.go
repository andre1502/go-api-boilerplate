package server

import (
	"errors"
	"go-api-boilerplate/internal/exception"
	"go-api-boilerplate/internal/status_code"
	"net/http"
	"reflect"
	"runtime"

	"github.com/labstack/echo/v5"
)

func (s *Server) CustomErrorHandler(c *echo.Context, err error) {
	var ex *exception.Exception
	var sc echo.HTTPStatusCoder
	var he *echo.HTTPError
	var ta *runtime.TypeAssertionError
	var message any

	httpStatus := http.StatusInternalServerError
	code := status_code.SERVER_ERROR_CODE
	message = status_code.SERVER_ERROR_MESSAGE

	if errors.As(err, &ex) {
		httpStatus = http.StatusOK
		code = ex.Code
		message = ex.Message
	} else if errors.As(err, &he) {
		httpStatus = he.Code
		message = he.Message

		if internalErr := errors.Unwrap(he); internalErr != nil {
			if errors.As(internalErr, &he) {
				httpStatus = he.Code
				message = he.Message
			}
		}
	} else if errors.As(err, &sc) {
		httpStatus = sc.StatusCode()
		code = httpStatus
		message = err.Error()
	} else if errors.As(err, &ta) {
		httpStatus = http.StatusBadRequest
		code = status_code.TYPE_ASSERTION_ERROR_CODE
		message = status_code.TYPE_ASSERTION_ERROR_MESSAGE
		err = &CustomError{
			Message: ta.Error(),
			Err:     ta,
		}
	} else {
		valueErr := reflect.ValueOf(err)

		err = &CustomError{
			Message: err.Error(),
			Err:     s.getErrorFromPointer(valueErr),
		}
	}

	if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
		if resp.Committed {
			return // response has been already sent to the client by handler or some middleware
		}
	}

	if httpStatus == http.StatusNotFound {
		s.response.NotFound(c)
	} else {
		s.response.ErrorHandler(c, httpStatus, exception.Ex.Errors(code, message.(string), err))
	}
}

func (s *Server) getErrorFromPointer(val reflect.Value) error {
	if val.Kind() == reflect.Pointer {
		val = val.Elem()

		return s.getErrorFromPointer(val)
	}

	return errors.New(val.String())

}
