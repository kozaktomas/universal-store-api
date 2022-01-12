package main

import (
	"fmt"
	"github.com/kozaktomas/universal-store-api/config"
	"net/mail"
	"time"
)

// ValidateServiceNames checks for reserved service names
func ValidateServiceNames(names []string) error {
	reservedNames := map[string]bool{
		"metrics":   true,
		"log_level": true,
	}

	for _, name := range names {
		_, found := reservedNames[name]
		if found {
			return fmt.Errorf("could not use reserved serice name: %q", name)
		}
	}

	return nil
}

func Validate(field config.FieldConfig, value interface{}, valueSet bool) error {
	if field.Required != nil && *field.Required && !valueSet {
		return fmt.Errorf("field needs to be set")
	}

	switch value.(type) {
	case string:
		return validateString(field, value.(string))
	case float64:
		return validateNumber(field, value.(float64), valueSet)
	}

	panic("should never happen")
}

func validateNumber(field config.FieldConfig, value float64, valueSet bool) error {
	fieldType, err := field.GetType()
	if err != nil {
		return err
	}

	if fieldType == config.FieldTypeInt {

		// check if float64 is actually int
		if value != float64(int(value)) {
			return fmt.Errorf("int expected, float given")
		}

		intValue := int(value)

		// check min value
		if valueSet && field.Min != nil && intValue < *field.Min {
			return fmt.Errorf("minimum is %d", *field.Min)
		}

		// check max value
		if valueSet && field.Max != nil && intValue > *field.Max {
			return fmt.Errorf("maximum is %d", *field.Max)
		}

		return nil
	}

	if fieldType == config.FieldTypeFloat {
		return nil
	}

	return fmt.Errorf("invalid type of string field %q", field.Type)
}

func validateString(field config.FieldConfig, value string) error {
	fieldType, err := field.GetType()
	if err != nil {
		return err
	}

	length := len(value)

	if fieldType == config.FieldTypeString {

		// required field
		if field.Required != nil && *field.Required {
			if length == 0 {
				return fmt.Errorf("field needs to be set")
			}
		}

		// min length field
		if length > 0 && field.Min != nil && *field.Min > 0 {
			if length < *field.Min {
				return fmt.Errorf("min '%d' lenght required", *field.Min)
			}
		}

		// max length field
		if length > 0 && field.Max != nil && *field.Max > 0 {
			if length > *field.Max {
				return fmt.Errorf("max '%d' lenght required", *field.Max)
			}
		}

		// check rules
		if field.Rule != nil && len(*field.Rule) > 0 {

			// email rule
			if length > 0 && field.GetRule() == config.FieldRuleEmail {
				_, err = mail.ParseAddress(value)
				if err != nil {
					return fmt.Errorf("valid email address required")
				}
			}
		}
		return nil
	}

	if fieldType == config.FieldTypeDate {

		// required field
		if field.Required != nil && *field.Required {
			if length == 0 {
				return fmt.Errorf("field needs to be set")
			}
		}

		// format
		if field.Format != nil && *field.Format == "" {
			return fmt.Errorf("format field is required for date type")
		}

		if field.Format != nil && length > 0 {
			_, err := time.Parse(*field.Format, value)
			if err != nil {
				return fmt.Errorf("could not parse date %q using format %q", value, *field.Format)
			}
		}

		return nil
	}

	return fmt.Errorf("invalid type of string field %q", field.Type)
}
