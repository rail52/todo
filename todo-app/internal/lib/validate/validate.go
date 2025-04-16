package validate

import (
	"github.com/go-playground/validator/v10"
)

var valid = validator.New()

func IsValid(i interface{}) error {
	return valid.Struct(i)
}
