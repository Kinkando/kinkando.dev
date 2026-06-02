package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var v *validator.Validate

func init() {
	v = validator.New(validator.WithRequiredStructEnabled())

	// Use the json tag name in error messages so field names match the JSON API
	// (e.g. "meal_type" instead of "MealType").
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
}

// Struct validates s against its `validate` struct tags and returns a
// *fiber.Error (400) with a human-readable message on the first failing
// constraint. Returns nil if all constraints pass.
func Struct(s any) error {
	if err := v.Struct(s); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok && len(ve) > 0 {
			return fiber.NewError(fiber.StatusBadRequest, fieldMessage(ve[0]))
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return nil
}

// fieldMessage converts a single FieldError into a user-facing message.
func fieldMessage(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required", "required_if":
		return fmt.Sprintf("%s is required", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(fe.Param(), " ", ", "))
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "gte", "min":
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "lte", "max":
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
