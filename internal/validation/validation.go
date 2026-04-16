package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-api-boilerplate/internal/status_code"
	"go-api-boilerplate/module"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

type Validation struct {
	validator *validator.Validate
}

func NewValidation() *Validation {
	validation := &Validation{
		validator: validator.New(),
	}

	validation.validator.RegisterTagNameFunc(validation.GetJsonTagName())
	validation.validator.RegisterValidation("empty_string", validation.EmptyString())
	validation.validator.RegisterValidation("is_email", validation.IsEmail())
	validation.validator.RegisterValidation("is_password_complex", validation.IsPasswordComplex())

	return validation
}

func (v *Validation) GetJsonTagName() func(fld reflect.StructField) string {
	return func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	}
}

func (v *Validation) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return err.(validator.ValidationErrors)
	}

	return nil
}

func (v *Validation) ValidateRequest(c *echo.Context, i interface{}) error {
	if err := c.Bind(i); err != nil {
		return err
	}

	if err := c.Validate(i); err != nil {
		return err
	}

	return nil
}

func (v *Validation) FormatValidationErrors(err error) []string {
	var validationErrors validator.ValidationErrors
	var httpError *echo.HTTPError
	var jsonUnmarshalErr *json.UnmarshalTypeError
	var strconvErr *strconv.NumError

	errorList := make([]string, 0)

	if errors.As(err, &validationErrors) {
		for _, e := range validationErrors {
			field := e.Field()
			tag := v.validationTagMap(e.Tag())
			param := e.Param()

			// 构建错误消息格式: <field>:<validation>:<value>
			errorMsg := field + ":" + tag
			if !module.IsEmptyString(param) {
				errorMsg += ":" + param
			}

			errorList = append(errorList, errorMsg)
		}
	} else if errors.As(err, &httpError) {
		if internalErr := errors.Unwrap(httpError); internalErr != nil {
			if errors.As(internalErr, &jsonUnmarshalErr) {
				errMsg := fmt.Sprintf("%s:is_%s", jsonUnmarshalErr.Field, jsonUnmarshalErr.Type)
				errorList = append(errorList, errMsg)
			} else if errors.As(internalErr, &strconvErr) {
				errMsg := strconvErr.Num

				switch strconvErr.Func {
				case "ParseInt":
					errMsg += ":is_not_number"
				case "ParseUint":
					errMsg += ":is_not_unsigned_number"
				case "ParseBool":
					errMsg += ":is_not_boolean"
				case "ParseFloat":
					errMsg += ":is_not_float"
				default:
					errMsg = strconvErr.Error()
				}

				errorList = append(errorList, errMsg)
			} else {
				errorList = append(errorList, internalErr.Error())
			}
		} else {
			errorList = append(errorList, httpError.Error())
		}
	} else {
		errorList = append(errorList, status_code.UNKNOWN_ERROR_MESSAGE)
	}

	return errorList
}

func (v *Validation) validationTagMap(tag string) string {
	switch tag {
	case "eqfield":
		return "must_equal"
	}

	return tag
}
