package response

import (
	"go-api-boilerplate/internal"
	"go-api-boilerplate/internal/exception"
	"go-api-boilerplate/internal/status_code"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/logger"
	"go-api-boilerplate/module/pagination"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Response struct {
	Pagination  *pagination.Pagination `json:"-"`
	Err         error                  `json:"-"`
	HttpStatus  int                    `json:"httpStatus"`
	Code        int                    `json:"code"`
	Message     string                 `json:"message"`
	Data        any                    `json:"data,omitempty"`
	Validations any                    `json:"validations,omitempty"`
}

type DataWithPagination struct {
	Items any `json:"items"`
	*pagination.Pagination
}

func NewResponse(pgn *pagination.Pagination) *Response {
	return &Response{
		Pagination: pgn,
		Data:       nil,
	}
}

func (r *Response) json(c echo.Context, httpStatus int, data any, err error) error {
	requestUri := c.Request().RequestURI
	r.HttpStatus = httpStatus
	r.Data = nil
	r.Validations = nil

	if err != nil {
		if he, ok := err.(*exception.Exception); ok {
			r.Code = he.Code
			r.Message = he.Error()
			r.Err = he.Err
		} else {
			r.Err = err
		}

		logData := logger.GetLogFields(c.Request().Method, requestUri, c.Response().Header().Get(echo.HeaderXRequestID), module.GetClientIP(c.Request()),
			r.Err, r.HttpStatus, r.Code, r.Message)

		logData["elastic_index"] = internal.GetLogIndex(requestUri)
		logger.Log.WithFields(logData).Error(err)

		return c.JSONPretty(r.HttpStatus, r, "  ")
	}

	r.Code = status_code.SUCCESS_CODE
	r.Message = status_code.SUCCESS_MESSAGE

	pagination := r.Pagination.PaginationFromContext(c.Request().Context())

	if (pagination != nil) && (pagination.Page > 0) && (pagination.TotalRecord > 0) {
		r.Data = &DataWithPagination{
			Items:      data,
			Pagination: pagination,
		}
	} else if module.IsArray(data) {
		r.Data = &DataWithPagination{
			Items: data,
		}
	} else {
		r.Data = data
	}

	return c.JSONPretty(r.HttpStatus, r, "  ")
}

func (r *Response) Success(c echo.Context, data any) error {
	return r.json(c, http.StatusOK, data, nil)
}

func (r *Response) Failed(c echo.Context, err error) error {
	return r.json(c, http.StatusOK, nil, err)
}

func (r *Response) ValidationFailed(c echo.Context, err *exception.Exception) error {
	var validationErrors []string
	if len(err.Args) > 0 {
		if errors, ok := err.Args[0].([]string); ok {
			validationErrors = errors
		}
	}
	r.HttpStatus = http.StatusBadRequest
	r.Code = err.Code
	r.Message = err.Message
	r.Data = nil
	r.Validations = validationErrors

	return c.JSONPretty(http.StatusBadRequest, r, "  ")
}

func (r *Response) Unauthorized(c echo.Context) error {
	return c.NoContent(http.StatusUnauthorized)
}

func (r *Response) NotFound(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}

func (r *Response) ErrorHandler(c echo.Context, httpStatus int, err error) error {
	return r.json(c, httpStatus, nil, err)
}
