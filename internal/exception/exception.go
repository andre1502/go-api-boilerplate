package exception

import (
	"fmt"
	"go-api-boilerplate/module/logger"
)

var Ex *Exception

type Exception struct {
	Code    int
	Message string
	Args    []interface{}
	Err     error
}

func NewException() *Exception {
	Ex = &Exception{}

	return Ex
}

func (ex *Exception) Errors(code int, message string, err error, args ...interface{}) *Exception {
	ex.Code = code
	ex.Message = message
	ex.Args = args
	ex.Err = err

	return ex
}

func (ex *Exception) Error() string {
	return ex.GetErrorMessage(ex.Code, ex.Message, ex.Args...)
}

func (ex *Exception) GetErrorMessage(code int, message string, args ...interface{}) string {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	return message
}

func (ex *Exception) ValidationError(code int, message string, validationErrors []string) *Exception {
	ex.Code = code
	ex.Message = message
	ex.Args = []interface{}{validationErrors}
	ex.Err = nil

	logger.Log.Error(ex)

	return ex
}
