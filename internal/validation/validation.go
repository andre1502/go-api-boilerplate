package validation

import (
	"encoding/json"
	"fmt"
	"go-api-boilerplate/internal/status_code"
	"go-api-boilerplate/module"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
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

func (v *Validation) ValidateRequest(c echo.Context, i interface{}) error {
	binder := echo.DefaultBinder{}
	if err := binder.BindQueryParams(c, i); err != nil {
		return err
	}

	if err := c.Bind(i); err != nil {
		return err
	}

	if err := c.Validate(i); err != nil {
		return err
	}

	return nil
}

func (v *Validation) FormatValidationErrors(err error) []string {
	errors := make([]string, 0)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := v.validationTagMap(e.Tag())
			param := e.Param()

			// validation error message format: <field>:<validation>:<value>
			errorMsg := field + ":" + tag
			if !module.IsEmptyString(param) {
				errorMsg += ":" + param
			}

			errors = append(errors, errorMsg)
		}
	} else if httpError, ok := err.(*echo.HTTPError); ok {
		if httpError.Internal != nil {
			if jsonUnmarshalErr, ok := httpError.Internal.(*json.UnmarshalTypeError); ok {
				errMsg := fmt.Sprintf("%s:is_%s", jsonUnmarshalErr.Field, jsonUnmarshalErr.Type)
				errors = append(errors, errMsg)
			} else if strconvErr, ok := httpError.Internal.(*strconv.NumError); ok {
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

				errors = append(errors, errMsg)
			} else {
				errors = append(errors, httpError.Internal.Error())
			}
		} else {
			errors = append(errors, httpError.Error())
		}
	} else {
		errors = append(errors, status_code.UNKNOWN_ERROR_MESSAGE)
	}

	return errors
}

func (v *Validation) validationTagMap(tag string) string {
	switch tag {
	case "eqfield":
		return "must_equal"
	}

	return tag
}
