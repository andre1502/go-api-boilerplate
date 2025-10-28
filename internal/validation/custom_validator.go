package validation

import (
	"go-api-boilerplate/module"
	"net/mail"
	"regexp"

	"github.com/go-playground/validator/v10"
)

func (v *Validation) EmptyString() func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		if str, ok := fl.Field().Interface().(string); ok {
			if !module.IsEmptyString(str) {
				return true
			}
		}

		return false
	}
}

func (v *Validation) IsEmail() func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		if str, ok := fl.Field().Interface().(string); ok {
			email, err := mail.ParseAddress(str)

			return err == nil && email.Address == str
		}

		return false
	}
}

// Pre-compile the regex patterns once at initialization (like in init() or as package variables).
// This fixes the repeated compilation issue in the loop.
var compiledPasswordRegex = []*regexp.Regexp{
	regexp.MustCompile(`.{8,}`),         // Minimum 8 characters
	regexp.MustCompile(`[a-z]`),         // At least one lowercase
	regexp.MustCompile(`[A-Z]`),         // At least one uppercase
	regexp.MustCompile(`[0-9]`),         // At least one digit
	regexp.MustCompile(`[^a-zA-Z\d\s]`), // At least one special character
}

func (v *Validation) IsPasswordComplex() func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		if password, ok := fl.Field().Interface().(string); ok {
			if module.IsEmptyString(password) {
				return false
			}

			for _, rx := range compiledPasswordRegex {
				// Use the compiled object's MatchString method.
				if !rx.MatchString(password) {
					return false // Fail fast on the first mismatch
				}
			}

			return true
		}

		return false
	}
}
