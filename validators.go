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
	// field required, not set
	if field.Required != nil && *field.Required && !valueSet {
		return fmt.Errorf("field %q: required", field.Name)
	}

	// field not required, not set
	if field.Required != nil && !*field.Required && !valueSet {
		return nil
	}

	var err error
	fieldType, err := field.GetType()
	if err != nil {
		return fmt.Errorf("filed %q: unknown type", field.Name)
	}

	switch fieldType {
	case config.FieldTypeObject:
		v, converted := value.(map[string]interface{})
		if !converted {
			return fmt.Errorf("field %q: could not expand object", field.Name)
		}
		for n, f := range *field.Fields {
			if err := Validate(*f, v[n], v[n] != nil); err != nil {
				return err
			}
		}

		return nil
	case config.FieldTypeString:
		strValue, converted := value.(string)
		if !converted {
			return fmt.Errorf("field %q: could not convert to string", field.Name)
		}
		return validateString(field, strValue)
	case config.FieldTypeDate:
		strValue, converted := value.(string)
		if !converted {
			return fmt.Errorf("field %q: could not convert to date", field.Name)
		}
		return validateDate(field, strValue)
	case config.FieldTypeInt:
		floatValue, converted := value.(float64)
		if !converted {
			return fmt.Errorf("field %q: could not convert to int", field.Name)
		}
		return validateInt(field, floatValue)
	case config.FieldTypeFloat:
		floatValue, converted := value.(float64)
		if !converted {
			return fmt.Errorf("field %q: could not convert to float", field.Name)
		}
		return validateFloat(field, floatValue)
	}

	panic(fmt.Sprintf("should never happen; field: %s", field.Name))
}

func validateString(field config.FieldConfig, value string) error {
	length := len(value)

	// required field
	if field.Required != nil && *field.Required {
		if length == 0 {
			return fmt.Errorf("field %q: required", field.Name)
		}
	}

	// min length field
	if field.Min != nil && *field.Min > 0 {
		if length < *field.Min {
			return fmt.Errorf("field %q: min '%d' lenght required", field.Name, *field.Min)
		}
	}

	// max length field
	if field.Max != nil && *field.Max > 0 {
		if length > *field.Max {
			return fmt.Errorf("field %q: max '%d' lenght required", field.Name, *field.Max)
		}
	}

	// check rules
	if field.Rule != nil && len(*field.Rule) > 0 {

		// email rule
		if length > 0 && field.GetRule() == config.FieldRuleEmail {
			_, err := mail.ParseAddress(value)
			if err != nil {
				return fmt.Errorf("field %q: valid email address required", field.Name)
			}
		}
	}

	return nil
}

func validateDate(field config.FieldConfig, value string) error {
	length := len(value)

	// required field
	if field.Required != nil && *field.Required {
		if length == 0 {
			return fmt.Errorf("field %q required", field.Name)
		}
	}

	// format
	if field.Format != nil && *field.Format == "" {
		return fmt.Errorf("field %q: format is required for date type", field.Name)
	}
	if field.Format != nil && length > 0 {
		_, err := time.Parse(*field.Format, value)
		if err != nil {
			return fmt.Errorf("field %q: could not parse date %q using format %q", field.Name, value, *field.Format)
		}
	}

	return nil
}

func validateInt(field config.FieldConfig, value float64) error {
	if value != float64(int(value)) {
		return fmt.Errorf("field %q: could not convert float %f to int", field.Name, value)
	}

	intValue := int(value)

	// check min value
	if field.Min != nil && intValue < *field.Min {
		return fmt.Errorf("field %q: minimum is %d", field.Name, *field.Min)
	}

	// check max value
	if field.Max != nil && intValue > *field.Max {
		return fmt.Errorf("field %q: maximum is %d", field.Name, *field.Max)
	}

	return nil
}

func validateFloat(field config.FieldConfig, value float64) error {
	// check min value
	if field.Min != nil && value < float64(*field.Min) {
		return fmt.Errorf("field %q: minimum is %d", field.Name, *field.Min)
	}

	// check max value
	if field.Max != nil && value > float64(*field.Max) {
		return fmt.Errorf("field %q: maximum is %d", field.Name, *field.Max)
	}

	return nil
}
