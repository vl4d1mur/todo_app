package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		return err
	}
	return nil
}
