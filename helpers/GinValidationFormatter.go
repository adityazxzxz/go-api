package helpers

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func GinValidationFormatter(err error, obj interface{}) map[string][]string {
	errors := make(map[string][]string)
	requiredFields := make(map[string]bool)

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			field := getJSONFieldName(obj, fe.Field())

			if fe.Tag() == "required" {
				errors[field] = []string{field + " is required"}
				requiredFields[field] = true
				continue
			}

			if requiredFields[field] {
				continue
			}

			var msg string
			switch fe.Tag() {
			case "email":
				msg = field + " must be a valid email"
			case "min":
				msg = field + " is too short"
			case "max":
				msg = field + " is too long"
			default:
				msg = field + " is invalid"
			}

			errors[field] = append(errors[field], msg)
		}
	}

	return errors
}

func getJSONFieldName(obj interface{}, fieldName string) string {
	t := reflect.TypeOf(obj)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	field, ok := t.FieldByName(fieldName)
	if !ok {
		return strings.ToLower(fieldName)
	}

	tag := field.Tag.Get("json")
	if tag == "" {
		return strings.ToLower(fieldName)
	}

	return strings.Split(tag, ",")[0]
}
