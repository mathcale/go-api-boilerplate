package handlers

import (
	"strconv"

	"github.com/go-playground/validator/v10"
)

func newRequestValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	// go-playground's max counts runes, but bcrypt's limit is in bytes
	_ = v.RegisterValidation("maxbytes", func(fl validator.FieldLevel) bool {
		limit, err := strconv.Atoi(fl.Param())
		if err != nil {
			return false
		}

		return len(fl.Field().String()) <= limit
	})

	return v
}
